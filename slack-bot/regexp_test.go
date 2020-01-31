package slack_bot

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var text = "<@H7BK8G35E> someshit <@C8BK7G35E>"

func Test_userName_regexp(t *testing.T) {

	names := userNameRegexp.FindAllString(text, -1)
	assert.Len(t, names, 2)
	assert.Contains(t, names, "<@H7BK8G35E>")
	assert.Contains(t, names, "<@C8BK7G35E>")
}

func Test_replaceUserNames(t *testing.T) {

	namesMap := map[string]string{
		"<@C8BK7G35E>": "Testator",
		"<@H7BK8G35E>": "Ololosh",
	}
	result := replaceUserNames(text, namesMap)

	assert.Equal(t, "Ololosh someshit Testator", result)
}
