package whatsappclient

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	whatsapp "github.com/Rhymen/go-whatsapp"
	"github.com/Rhymen/go-whatsapp/binary/proto"
)

type messageListener func(Message)

type messageHandler struct {
	listener  messageListener
	afterTime int64
}

func (h *messageHandler) HandleError(err error) {
	fmt.Fprintf(os.Stderr, "%v", err)
}

func (h *messageHandler) HandleTextMessage(message whatsapp.TextMessage) {
	// if message.Info.FromMe {
	// 	return
	// }

	msgTime := int64(message.Info.Timestamp)

	if h.afterTime > msgTime {
		return
	}

	mapMessage := Message{
		Type:       "text",
		ID:         message.Info.Id,
		From:       message.Info.RemoteJid,
		Info:       message.Info,
		TextSource: message,
	}

	h.listener(mapMessage)
}

func (h *messageHandler) HandleImageMessage(message whatsapp.ImageMessage) {
	// if message.Info.FromMe {
	// 	return
	// }

	msgTime := int64(message.Info.Timestamp)

	if h.afterTime > msgTime {
		return
	}

	mapMessage := Message{
		Type:        "image",
		ID:          message.Info.Id,
		From:        message.Info.RemoteJid,
		Info:        message.Info,
		ImageSource: message,
	}

	h.listener(mapMessage)
}

func (h *messageHandler) HandleStickerMessage(message whatsapp.StickerMessage) {
	// if message.Info.FromMe {
	// 	return
	// }

	msgTime := int64(message.Info.Timestamp)

	if h.afterTime > msgTime {
		return
	}

	mapMessage := Message{
		Type:          "sticker",
		ID:            message.Info.Id,
		From:          message.Info.RemoteJid,
		Info:          message.Info,
		StickerSource: message,
	}

	h.listener(mapMessage)
}

// func (h *messageHandler) HandleJsonMessage(message string) {
// 	// if message.Info.FromMe {
// 	// 	return
// 	// }
// 	fmt.Println("JSON : ", message)
// 	// msgTime := int64(message.Info.Timestamp)

// 	// if h.afterTime > msgTime {
// 	// 	return
// 	// }

// 	// mapMessage := Message{
// 	// 	Type:          "sticker",
// 	// 	ID:            message.Info.Id,
// 	// 	From:          message.Info.RemoteJid,
// 	// 	Info:          message.Info,
// 	// 	StickerSource: message,
// 	// }

// 	// h.listener(mapMessage)
// }

func (h *messageHandler) HandleRawMessage(message *proto.WebMessageInfo) {

	if message.MessageTimestamp == nil {
		return
	}
	if int64(*message.MessageTimestamp) < h.afterTime {
		return
	}
	_, err := json.Marshal(message)
	//fmt.Println("RAW : ", b)
	if err != nil {
		fmt.Println(err)
		return
	}
}
func (h *messageHandler) HandleJsonMessage(message string) {
	//fmt.Println(message)
	var (
		groupID  string
		action   string
		actionBy string
		part     map[string]interface{}
	)
	var array []interface{}
	dec := json.NewDecoder(strings.NewReader(message))
	err := dec.Decode(&array)
	if err == nil {
		if len(array) > 0 {
			if array[0] == "Chat" {
				// save group activity (user join/leave)
				data, ok := array[1].(map[string]interface{})
				if ok {
					groupID = data["id"].(string)
					if data["cmd"] == "action" {
						var chatData []interface{}
						chatData = data["data"].([]interface{})
						action = chatData[0].(string)
						actionBy = chatData[1].(string)
						part, _ = chatData[2].(map[string]interface{})
						//fmt.Println(groupID, chatData, action, by, part)
						// if ok {
						// 	if partInterface, ok := part["participants"].([]interface{}); ok {
						// 		for _, v := range partInterface {
						// 			activity := database.GroupActivity{GroupID: groupID, Action: action, By: by, UID: strings.Replace(v.(string), "@c.us", "@s.whatsapp.net", 1)}
						// 			err := conn.InsertGroupActivity(&activity)
						// 			if err != nil {
						// 				log.Println(err)
						// 			}
						// 		}
						// 	}
						// }
					}
				}
			}
		}
	}
	if groupID != "" {
		mapMessage := Message{
			Type: "action",
			From: groupID,
			Actions: GroupAction{
				Action:       action,
				ActionBy:     actionBy,
				Participants: part,
			},
		}

		h.listener(mapMessage)
	}
}
