package whatsappclient

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"time"

	whatsapp "github.com/Rhymen/go-whatsapp"
	"github.com/disintegration/imaging"
	qrcode "github.com/skip2/go-qrcode"
)

// WhatsappClient connect
type WhatsappClient struct {
	wac *whatsapp.Conn
}

type ThumbUrl struct {
	EURL   string `json:"eurl"`
	Tag    string `json:"tag"`
	Status int64  `json:"status"`
}

func NewClient() *WhatsappClient {
	fmt.Println("Baymax Bot - A whatsapp management bot")
	wac, err := whatsapp.NewConn(5 * time.Minute)
	wac.SetClientName("Baymax Bot Web API", "Whatsapp API", "0.6.600")
	wac.SetClientVersion(0, 6, 600)
	if err != nil {
		fmt.Println(err)
	}
	err = loginToWhatsapp(wac)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error logging in: %v\n", err)
		os.Exit(1)
		return nil
	}

	fmt.Println("Baymax Bot Activated ...")
	return &WhatsappClient{wac}

}

func (wp *WhatsappClient) Listen(f messageListener) {
	wp.wac.AddHandler(&messageHandler{f, time.Now().Unix()})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	fmt.Println("Press ctrl+c to exit.")

	<-sigs
	fmt.Println("Shutdown.")
	os.Exit(0)
}

//Send name of user

func (wp *WhatsappClient) GetProfileImage(userID string) (string, string) {

	// RemoteJID is the sender ID the one who send the text or the media message or basically the WhatsappID just pass it here
	profilePicThumb, _ := wp.wac.GetProfilePicThumb(userID)
	profilePic := <-profilePicThumb

	thumbnail := ThumbUrl{}
	err := json.Unmarshal([]byte(profilePic), &thumbnail)
	fmt.Println(thumbnail)
	if err != nil {
		return "", ""
	}
	if thumbnail.Status == 404 {
		// meaning thumbnail is not available because the person has no profile pic so what i did is return empty string
		return "", ""
	}
	// Basically the EURL is what holds the profile picture of the person
	return thumbnail.EURL, thumbnail.Tag
}

func (wp *WhatsappClient) GetName(userID string) string {
	name := wp.wac.Store.Contacts[userID].Notify
	return name
}

func containID(data []string, ID string) bool {
	for _, i := range data {
		if i == ID {
			return true
		}
	}
	return false
}

//Delete message for everyone
func (wp *WhatsappClient) Purge(to string, messageID string) {
	file, err := ioutil.ReadFile("bot-data/sent-message-ids.json")
	if err != nil {
		fmt.Println("Purge Error : JSON Read file error")
	} else {
		var obj toJSON
		json.Unmarshal(file, &obj)
		for index, data := range obj.Data {
			if data.Origin == to && containID(data.ID, messageID) {
				for i := len(data.ID) - 1; i >= 0; i-- {
					if data.ID[i] == messageID {
						msgID, err := wp.wac.RevokeMessage(to, data.ID[i], true)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error deleting message: %v", err)
						} else {
							fmt.Println(msgID, " Message Purged.")
							obj.Data[index].ID = obj.Data[index].ID[:i]
						}
						break
					}
					msgID, err := wp.wac.RevokeMessage(to, data.ID[i], true)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error deleting message: %v", err)
					} else {
						fmt.Println(msgID, " Message Purged.")
						obj.Data[index].ID = obj.Data[index].ID[:i]
					}
				}
				dataJSON, err := json.Marshal(obj)
				if err != nil {
					fmt.Println("Failed to Remove ID/s from log, marshal error")
					return
				}
				//fmt.Println(string(libJSON))
				if err := ioutil.WriteFile("bot-data/sent-message-ids.json", dataJSON, 0644); err != nil {
					fmt.Println("Failed to ADD ID to log, Write file error")
				}
				break
			}

		}
	}

}

//Delete Message. It is not useful in group.
// func (wp *WhatsappClient) DeleteText(to string, messageID string, fromMe bool) {

// 	err := wp.wac.DeleteMessage(to, messageID, fromMe)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Error deleting message: %v", err)
// 	} else {
// 		fmt.Println("Message Deleted.")
// 	}
// }

// Source: &proto.WebMessageInfo{
// 	Message: &proto.Message{
// 		ExtendedTextMessage: &proto.ExtendedTextMessage{
// 			Text: &text,
// 			ContextInfo: &proto.ContextInfo{
// 				MentionedJid: []string{mid},
// 			},
// 			PreviewType: nil,
// 		},
// 	},
// },

type sentMessageID struct {
	Origin string
	ID     []string
}

type toJSON struct {
	Data []sentMessageID
}

func addIDtoLog(msgID string, msgFrom string) {
	f, err := ioutil.ReadFile("bot-data/sent-message-ids.json")
	var obj toJSON
	if err != nil {
		fmt.Println("Failed to read from log file, read file error")
		obj.Data = append(obj.Data, sentMessageID{Origin: msgFrom, ID: []string{msgID}})
	} else {
		err = json.Unmarshal(f, &obj)
		if err != nil {
			fmt.Println("Failed to ADD ID to log, Unmarshal error")
			return
		}
		isFound := false
		for index, data := range obj.Data {
			if data.Origin == msgFrom {
				obj.Data[index].ID = append(obj.Data[index].ID, msgID)
				isFound = true
				break
			}
		}
		if !isFound {
			obj.Data = append(obj.Data, sentMessageID{Origin: msgFrom, ID: []string{msgID}})
		}
	}
	dataJSON, err := json.Marshal(obj)
	if err != nil {
		fmt.Println("Failed to ADD ID to log, marshal error")
		return
	}
	//fmt.Println(string(libJSON))
	if err := ioutil.WriteFile("bot-data/sent-message-ids.json", dataJSON, 0644); err != nil {
		fmt.Println("Failed to write to log file, Write file error")
	}

}

//Send Text
func (wp *WhatsappClient) SendText(to string, text string, msgSource whatsapp.MessageInfo, isReply bool) {
	var reply whatsapp.TextMessage
	if isReply && !msgSource.FromMe && !strings.Contains(to, "@s.whatsapp.net") {
		reply = whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: to,
				Source:    msgSource.Source,
			},
			Text: text,
			ContextInfo: whatsapp.ContextInfo{
				QuotedMessageID: msgSource.Id,
				QuotedMessage:   msgSource.Source.Message,
				Participant:     *msgSource.Source.Participant,
			},
		}
	} else {
		reply = whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: to,
			},
			Text: text,
		}
	}

	msgID, err := wp.wac.Send(reply)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending message: %v", err)
	} else {
		fmt.Println("Message Sent -> ID : " + msgID)
		addIDtoLog(msgID, to)
	}
}

func getThumbnail(images io.Reader) ([]byte, error) {
	img, _, err := image.Decode(images)
	if err != nil {
		return nil, err
	}

	b := img.Bounds()
	imgWidth := b.Max.X
	imgHeight := b.Max.Y

	thumbWidth := 100
	thumbHeight := 100

	if imgWidth > imgHeight {
		thumbHeight = 56
	} else {
		thumbWidth = 56
	}

	thumb := imaging.Thumbnail(img, thumbWidth, thumbHeight, imaging.CatmullRom)

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, thumb, nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Send Image
func (wp *WhatsappClient) SendImage(to string, text string, imageType string, content []byte, msgSource whatsapp.MessageInfo, isReply bool) {
	thumb, err := getThumbnail(bytes.NewReader(content))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var reply whatsapp.ImageMessage
	if isReply && !msgSource.FromMe && !strings.Contains(to, "@s.whatsapp.net") {
		reply = whatsapp.ImageMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: to,
				Source:    msgSource.Source,
			},
			Type:      imageType,
			Content:   bytes.NewReader(content),
			Caption:   text,
			Thumbnail: thumb,
			ContextInfo: whatsapp.ContextInfo{
				QuotedMessageID: msgSource.Id,
				QuotedMessage:   msgSource.Source.Message,
				Participant:     *msgSource.Source.Participant,
			},
		}
	} else {
		reply = whatsapp.ImageMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: to,
			},
			Type:      imageType,
			Content:   bytes.NewReader(content),
			Caption:   text,
			Thumbnail: thumb,
		}
	}
	msgID, err := wp.wac.Send(reply)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending message: %v", err)
	} else {
		fmt.Println("Message Sent -> ID : " + msgID)
		addIDtoLog(msgSource.Id, to)
	}
}

// Send Sticker
func (wp *WhatsappClient) SendSticker(to string, content io.Reader, msgSource whatsapp.MessageInfo, isReply bool) {
	var reply whatsapp.StickerMessage
	if isReply && !msgSource.FromMe && !strings.Contains(to, "@s.whatsapp.net") {
		reply = whatsapp.StickerMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: to,
				Source:    msgSource.Source,
			},
			Type:    "image/webp",
			Content: content,
			ContextInfo: whatsapp.ContextInfo{
				QuotedMessageID: msgSource.Id,
				QuotedMessage:   msgSource.Source.Message,
				Participant:     *msgSource.Source.Participant,
			},
		}
	} else {
		reply = whatsapp.StickerMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: to,
			},
			Type:    "image/webp",
			Content: content,
		}
	}

	msgID, err := wp.wac.Send(reply)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending message: %v", err)
	} else {
		fmt.Println("Message Sent -> ID : " + msgID)
		addIDtoLog(msgSource.Id, to)
	}
}

// GetConnection return whatsapp connection
func (wp *WhatsappClient) GetConnection() *whatsapp.Conn {
	return wp.wac
}

func loginToWhatsapp(wac *whatsapp.Conn) error {
	//load saved session
	session, err := readSession()
	if err == nil {
		//restore session
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v", err)
		}
	} else {
		//no saved session -> regular login
		qrChan := make(chan string)
		fmt.Println("Scan QR code.")
		go func() {
			err := qrcode.WriteFile(<-qrChan, qrcode.Medium, 256, "qr.png")
			if err != nil {
				fmt.Println(err)
			}
			//show qr code or save it somewhere to scan
		}()

		session, err = wac.Login(qrChan)
		if err != nil {
			return fmt.Errorf("error during login: %v", err)
		}
		fmt.Println("QR code scanned sucessfully.")

	}

	err = writeSession(session)
	if err != nil {
		return fmt.Errorf("error saving session: %v", err)
	}
	fmt.Println("\nLogged in sucessfully to baymax bot.")
	return nil
}

func readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	currentWorkingDir, _ := os.Getwd()
	file, err := os.Open(currentWorkingDir + "/baymaxBot.session")
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

func writeSession(session whatsapp.Session) error {
	currentWorkingDir, _ := os.Getwd()
	file, err := os.Create(currentWorkingDir + "/baymaxBot.session")
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}
