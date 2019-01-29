package ninjam_bot

import (
	"bufio"
	"github.com/Ayvan/ninjam-chatbot/models"
	"github.com/luci/go-render/render"
	"github.com/sirupsen/logrus"
	"net"
	"runtime"
	"time"
)

type NinJamBot struct {
	keepAliveTicker    *time.Ticker
	toServerChan       chan []byte
	inAuthNow          bool
	sigChan            chan bool
	users              map[string]string
	anonymous          bool
	userName           string
	password           string
	host               string
	port               string
	messagesFromNinJam chan models.Message
	messagesToNinJam   chan string
	adminMessages      chan string

	onSuccessAuth func()
}

func NewNinJamBot(host, port, userName, password string, anonymous bool) *NinJamBot {
	return &NinJamBot{
		keepAliveTicker:    time.NewTicker(time.Second * 10),
		toServerChan:       make(chan []byte, 1000),
		sigChan:            make(chan bool, 1),
		users:              make(map[string]string),
		anonymous:          anonymous,
		userName:           userName,
		password:           password,
		host:               host,
		port:               port,
		messagesFromNinJam: make(chan models.Message, 1000),
		messagesToNinJam:   make(chan string, 1000),
		adminMessages:      make(chan string, 1000),
	}
}

func (n *NinJamBot) Host() string {
	return n.host
}

func (n *NinJamBot) Port() string {
	return n.port
}

func (n *NinJamBot) UserName() string {
	return n.userName
}

func (n *NinJamBot) Connect() {
f:
	for {
		select {
		case s := <-n.sigChan:
			n.sigChan <- s
			break f
		default:

			n.connect()
			// если коннект прервался - запустим таймаут перед реконнектом
			time.Sleep(time.Second * 5)
		}
	}
}

func (n *NinJamBot) Stop() {
	n.sigChan <- true
}

func (n *NinJamBot) IncomingMessages() <-chan models.Message {
	return n.messagesFromNinJam
}

func (n *NinJamBot) SendMessage(message string) {
	go func() {
		n.messagesToNinJam <- message
	}()
}

func (n *NinJamBot) SendAdminMessage(message string) {
	go func() {
		n.adminMessages <- message
	}()
}

func (n NinJamBot) Users() []string {
	users := []string{}
	for userName := range n.users {
		users = append(users, userName)
	}

	return users
}

func (n *NinJamBot) connect() {
	defer func() {
		logrus.Info("connect finished")
	}()

	conn, err := dialNinjamServer(n.host, n.port)

	for err != nil {
		select {
		case s := <-n.sigChan:
			n.sigChan <- s
			return
		default:
			logrus.Error("Ninjam connection error", err)
			logrus.Info("Retry connecting after 5 seconds...")

			// ошибка коннекта, пробуем снова через таймаут 5 секунд
			time.Sleep(time.Second * 10)
			conn, err = dialNinjamServer(n.host, n.port)
		}
	}

	returnChan := make(chan bool, 10)

	defer conn.Close()

	toServerErrorChan := make(chan bool, 1)

	go func() {
		for {
			select {
			case <-n.keepAliveTicker.C:
				// пока авторизуемся - тикер вырубаем
				if n.inAuthNow {
					continue
				}
				n.toServerChan <- []byte{models.ClientKeepaliveType, 0, 0, 0, 0}
			case <-toServerErrorChan:
				returnChan <- true
			case s := <-n.sigChan:
				returnChan <- true
				n.sigChan <- s
				return
			}
		}
	}()

	go n.sendToServer(conn, toServerErrorChan)

	// запускаем обработку сообщений, отправляемых в Ninjam чат
	go func() {
		for {
			select {
			case message := <-n.messagesToNinJam:
				n.sendChatMessage(message, models.MSG)
			case message := <-n.adminMessages:
				n.sendChatMessage(message, models.ADMIN)
			case s := <-n.sigChan:
				n.sigChan <- s
				// получена команда выйти из горутины
				return
			}
		}
	}()

	defer func() {
		returnChan <- true
	}()
	// блокирующая функция, если она вылетела - значит ошибка чтения коннекта, пробуем реконнект
	n.read(conn, returnChan)
}

func (n *NinJamBot) login(serverAuthChallenge *models.ServerAuthChallenge) (data []byte, err error) {
	var userName string
	if n.anonymous {
		userName = "anonymous:" + n.userName
	} else {
		userName = n.userName
	}
	authMessage := models.NewClientAuthUser(userName, n.password, serverAuthChallenge.HasAgreement(), serverAuthChallenge.Challenge)

	nm := models.NewNetMessage(models.ClientAuthUserType)

	nm.OutPayload = authMessage

	return nm.Marshal()
}

func dialNinjamServer(host, port string) (conn net.Conn, err error) {
	address := host + ":" + port

	logrus.Info("Connecting to Ninjam... ", address)

	dialer := &net.Dialer{
		KeepAlive: time.Hour * 24,
		Timeout:   time.Second * 10,
	}

	conn, err = dialer.Dial("tcp", address)

	if err != nil {
		return nil, err
	}

	logrus.Info("Successfully connected to ", address)

	return conn, nil
}

func (n *NinJamBot) sendChatMessage(message string, msgType string) {
	nm := models.NewNetMessage(models.ChatMessageType)

	cm := &models.ChatMessage{
		Command: []byte(msgType),
		Arg1:    []byte(message),
	}

	nm.OutPayload = cm

	msg, err := nm.Marshal()
	if err != nil {
		logrus.Error("Send message to ninjam marshal error:", err)
	}

	n.toServerChan <- msg
}

// WaitAuth block until auth completed
func (n *NinJamBot) WaitAuth() {
	for n.inAuthNow {
		time.Sleep(time.Millisecond)
	}
}

func (n *NinJamBot) OnSuccessAuth(f func()) {
	n.onSuccessAuth = f
}

func (n *NinJamBot) ChannelInit(name string) {
	channelInfo := &models.ClientSetChannelInfo{
		Channels: []models.ChannelInfo{
			{
				Name: name,
			},
		},
	}

	nm := models.NewNetMessage(models.ClientSetChannelInfoType)

	nm.OutPayload = channelInfo

	msg, err := nm.Marshal()
	if err != nil {
		logrus.Error("Send message to ninjam marshal error:", err)
	}

	n.toServerChan <- msg
}

func (n *NinJamBot) IntervalBegin(guid [16]byte, channelIndex uint8) {
	if n.inAuthNow {
		return
	}

	nm := models.NewNetMessage(models.ClientUploadIntervalBeginType)

	cm := &models.ClientUploadIntervalBegin{
		GUID:         guid,
		ChannelIndex: channelIndex,
	}
	nm.OutPayload = cm

	msg, err := nm.Marshal()
	if err != nil {
		logrus.Error("Send message to ninjam marshal error:", err)
	}

	n.toServerChan <- msg
}

func (n *NinJamBot) IntervalWrite(guid [16]byte, data []byte, flags uint8) {
	if n.inAuthNow {
		return
	}

	nm := models.NewNetMessage(models.ClientUploadIntervalWriteType)

	cm := &models.ClientUploadIntervalWrite{
		GUID:      guid,
		Flags:     flags,
		AudioData: data,
	}
	nm.OutPayload = cm

	msg, err := nm.Marshal()
	if err != nil {
		logrus.Error("Send message to ninjam marshal error:", err)
	}

	n.toServerChan <- msg
}

func (n *NinJamBot) read(conn net.Conn, returnChan chan bool) {
	defer func() {
		conn.Close()
		logrus.Info("Conection closed")
	}()

	logrus.Info("Started connect reader...")

	reader := bufio.NewReader(conn)
	readChan := make(chan []byte, 1)

	for {
		// в горутине запускаем чтение коннекта
		go func() {
			b := make([]byte, 5)
			length, err := reader.Read(b)

			if err != nil {
				logrus.Infof("Error reading: %s", err.Error())
				returnChan <- true
				return
			} else if length < 5 {
				logrus.Info("Error reading: read less than 5 bytes")
				returnChan <- true
				return
			}
			logrus.Info("Read from server: ", b)
			readChan <- b
			return
		}()

		select {
		case <-returnChan:
			returnChan <- true
			// получена команда выйти из горутины
			return
		case b := <-readChan:

			newMessage := [5]byte{}
			copy(newMessage[:], b[0:5])

			netMessage := models.NewInNetMessage(newMessage)

			// читаем данные сообщения в буфер равный его заявленной длине
			payload := make([]byte, netMessage.Length)
			bufLen, err := reader.Read(payload)
			if err != nil {
				logrus.Debug("Error reading:", err.Error())
				return
			}

			if netMessage.Length != uint32(bufLen) {
				logrus.Warning("Error: wrong payload length; buffLen=", bufLen, ", expected length=", netMessage.Length)
				return
			}

			err = netMessage.Unmarshal(payload)

			if err != nil {
				logrus.Error("Error when unmarshalling payload:", err)
			} else {
				if netMessage.InPayload != nil {
					logrus.Info(render.Render(netMessage.InPayload))
				}

				logrus.Info("Raw bytes:", render.Render(netMessage.RawData))

				go n.handle(netMessage)
			}
		}
	}
}

// получаем из канала ответы и пишем в сокет
func (n *NinJamBot) sendToServer(conn net.Conn, toServerErrorChan chan bool) {
	for {
		runtime.Gosched()
		res := <-n.toServerChan
		if len(res) < 200 {
			logrus.Info("Sending to server: ", res)
		}
		_, err := conn.Write(res)

		if err != nil {
			logrus.Error("Error writing sendToServer:", err.Error())
			toServerErrorChan <- true
			return
		}
	}
}

func (n *NinJamBot) handle(netMessage *models.NetMessage) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("Handle error: %s", r)
			return
		}
	}()
	switch netMessage.Type {
	case models.ServerAuthChallengeType:
		n.inAuthNow = true
		go func() {
			// через 10 секунд всё равно отключим режим авторизации, если даже не получим ответа - в крайнем случае
			// по тикеру пошлём KeepAlive и переконнектимся после ошибки отправки
			time.Sleep(time.Second * 10)
			n.inAuthNow = false
		}()
		serverAuthChallenge := netMessage.InPayload.(*models.ServerAuthChallenge)

		answer, err := n.login(serverAuthChallenge)

		if err != nil {
			logrus.Error("Error when logging in:", err)
			return
		}

		keepAlive, err := serverAuthChallenge.KeepAliveInterval()

		if err != nil {
			logrus.Error("Error when decode keep alive interval in:", err)
			return
		}

		n.keepAliveTicker = time.NewTicker(keepAlive)

		n.toServerChan <- answer
	case models.ServerAuthReplyType:
		serverAuthReply := netMessage.InPayload.(*models.ServerAuthReply)

		if serverAuthReply.Flag == 0x1 {
			logrus.Infof("Logged in succesfully: %s", string(serverAuthReply.ErrorMessage))

			if n.onSuccessAuth != nil {
				n.onSuccessAuth()
			}
		} else {
			logrus.Errorf("Login failed: %s", string(serverAuthReply.ErrorMessage))
		}
		n.inAuthNow = false
	case models.ServerUserInfoChangeNotifyType:
		serverUserInfo := netMessage.InPayload.(*models.ServerUserInfoChangeNotify)

		for _, userInfo := range serverUserInfo.UserInfos {
			if userInfo.Active == 0x1 {
				n.users[string(userInfo.Name)] = string(userInfo.Name)
			} else {
				delete(n.users, string(userInfo.Name))
			}
		}
		logrus.Infof("Users: %v", n.users)
	case models.ChatMessageType:
		chatMessage := netMessage.InPayload.(*models.ChatMessage)

		logrus.Infof("Chat message received: %s %s %s %s %s", chatMessage.Command, chatMessage.Arg1, chatMessage.Arg2, chatMessage.Arg3, chatMessage.Arg4)

		command := string(chatMessage.Command)

		switch command {
		case models.MSG:
			m := models.Message{
				Type: command,
				Name: string(chatMessage.Arg1),
				Text: string(chatMessage.Arg2),
			}
			n.messagesFromNinJam <- m
			logrus.Infof("%s said: %s", chatMessage.Arg1, chatMessage.Arg2)
		case models.JOIN:
			m := models.Message{
				Type: command,
				Name: string(chatMessage.Arg1),
			}
			n.messagesFromNinJam <- m
			logrus.Infof(" %s joined", chatMessage.Arg1)
		case models.PART:
			m := models.Message{
				Type: command,
				Name: string(chatMessage.Arg1),
			}
			n.messagesFromNinJam <- m
			logrus.Infof(" %s leaved", chatMessage.Arg1)
		}
	}
}
