package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	whatsapp "github.com/Rhymen/go-whatsapp"
	client "github.com/parth-javiya/baymax-bot/whatsappclient"
)

func main() {
	newClient := client.NewClient()
	newClient.Listen(func(msg client.Message) {
		//fmt.Println("Full Message : ", msg)
		//fmt.Println("Text : ", msg.TextSource)
		// fmt.Println("Image : ", msg.ImageSource)
		//919409302258-1609270381@g.us, 918200625135@s.whatsapp.net
		//msg.From == "919606748590-1591280536@g.us" ||
		if !(msg.From == "919606748590-1591280536@g.us" || msg.From == "919409302258-1609270381@g.us" || msg.From == "919409302258-1609146912@g.us" || msg.From == "917990244305-1609147323@g.us") {
			fmt.Println("Blocked!!")
			return
		}
		if msg.Type == "text" {
			textHandler(msg, newClient)
		} else if msg.Type == "image" {
			return
			imageHandler(msg, newClient)
		} else if msg.Type == "sticker" {
			return
			stickerHandler(msg, newClient)
		} else if msg.Type == "action" {
			actionHandler(msg, newClient)
		}
	})
}

func checkBannedWords(sentence string) bool {
	bannedWords := []string{"fuck", "mc", "bc"}
	for _, word := range bannedWords {
		if strings.Contains(sentence, word) {
			return true
		}
	}
	return false
}

func textHandler(msg client.Message, newClient *client.WhatsappClient) {
	var userData string

	if checkBannedWords(strings.ToLower(msg.TextSource.Text)) {
		newClient.SendText(msg.From, "Banned words not allowed. Admin can kick you.", msg.Info, true)
		return
	}

	if strings.ToLower(msg.TextSource.Text) == "/hi" {
		handleHiCommand(msg.TextSource, newClient)
	} else if strings.ToLower(msg.TextSource.Text) == "/image" {
		imgData, _ := ioutil.ReadFile("images/tk.jpg")
		newClient.SendImage(msg.TextSource.Info.RemoteJid, "Tanjiro Kamado", "image/jpeg", imgData, msg.Info, true)
	} else if strings.ToLower(msg.TextSource.Text) == "/sticker" {
		stickerData, _ := os.Open("stickers/" + strconv.Itoa(rand.Intn(33)+1) + ".webp")
		newClient.SendSticker(msg.From, stickerData, msg.Info, true)
	} else if strings.ToLower(msg.TextSource.Text) == "/tag" {
		newClient.SendText(msg.From, "Currently tag command is not available.", msg.Info, false)
	} else if strings.ToLower(msg.TextSource.Text) == "/del" {
		if msg.TextSource.ContextInfo.QuotedMessageID == "" {
			fmt.Println("Text : ", "Delete last one")
		} else {
			newClient.Purge(msg.From, msg.TextSource.ContextInfo.QuotedMessageID)
		}
	} else if strings.HasPrefix(msg.TextSource.Text, "/profile") {
		userData = strings.Replace(msg.TextSource.Text, "/profile", "", 1)
		userData = strings.TrimSpace(userData)
		handleProfileCommand(userData, msg, newClient)
	} else if strings.HasPrefix(msg.TextSource.Text, "/calc") {
		userData = strings.Replace(msg.TextSource.Text, "/calc", "", 1)
		userData = strings.TrimSpace(userData)
		handleCalcCommand(userData, msg, newClient)
	} else if strings.HasPrefix(msg.TextSource.Text, "/getnotelist") {
		newClient.SendText(msg.From, getNoteList(msg.From), msg.Info, true)
	} else if strings.HasPrefix(msg.TextSource.Text, "/get ") {
		userData = strings.Replace(msg.TextSource.Text, "/get ", "", 1)
		userData = strings.TrimSpace(userData)
		note, ok := getSavedNotes(userData, msg.From)
		if ok {
			newClient.SendText(msg.From, note, msg.Info, true)
		} else {
			newClient.SendText(msg.From, "Note you are looking for is not available.\nplease check /getnotelist", msg.Info, true)
		}
	} else if strings.HasPrefix(msg.TextSource.Text, "/delnote") {
		userData = strings.Replace(msg.TextSource.Text, "/delnote", "", 1)
		userData = strings.TrimSpace(userData)
		ok := deleteSavedNotes(userData, msg.From)
		if ok {
			newClient.SendText(msg.From, "*"+userData+"* note deleted.", msg.Info, true)
		} else {
			newClient.SendText(msg.From, "Failed to delete *"+userData+"* note.", msg.Info, true)
		}
	} else if strings.HasPrefix(msg.TextSource.Text, "/save") {
		userData = strings.Replace(msg.TextSource.Text, "/save", "", 1)
		userData = strings.TrimSpace(userData)
		if msg.TextSource.ContextInfo.QuotedMessageID == "" {
			noteData := strings.SplitN(userData, " ", 2)
			if len(noteData) < 2 {
				newClient.SendText(msg.From, "/save note-name message", msg.Info, true)
				return
			}

			errMsg, ok := saveNote(noteData[0], noteData[1], msg.From)
			if ok {
				newClient.SendText(msg.From, "*"+noteData[0]+"* note saved successfully", msg.Info, true)
			} else {
				newClient.SendText(msg.From, errMsg, msg.Info, true)
			}
		} else {
			saveText := *msg.TextSource.ContextInfo.QuotedMessage.Conversation
			errMsg, ok := saveNote(userData, saveText, msg.From)
			if ok {
				newClient.SendText(msg.From, "*"+userData+"* note saved successfully", msg.Info, true)
			} else {
				newClient.SendText(msg.From, errMsg, msg.Info, true)
			}
		}
	}
}

type Notes struct {
	NoteName string
	NoteText string
	From     string
}

func saveNote(noteName string, noteData string, from string) (string, bool) {
	f, err := ioutil.ReadFile("notes/data.json")
	var obj []Notes
	if err != nil {
		fmt.Println("Failed to load note file, read file error")
		obj = append(obj, Notes{NoteName: noteName, NoteText: noteData, From: from})
	} else {
		err = json.Unmarshal(f, &obj)
		if err != nil {
			fmt.Println("Failed to ADD note, Unmarshal error")
			return "Error occured.", false
		}

		isFound := false
		for _, data := range obj {
			if data.NoteName == noteName && data.From == from {
				return "Duplicate notes not allowed\ntry to save note with another name.", false
			}
		}
		if !isFound {
			obj = append(obj, Notes{NoteName: noteName, NoteText: noteData, From: from})
		}
	}
	dataJSON, err := json.Marshal(obj)
	if err != nil {
		fmt.Println("Failed to ADD ID to log, marshal error")
		return "Error occured.", false
	}
	//fmt.Println(string(libJSON))
	if err := ioutil.WriteFile("notes/data.json", dataJSON, 0644); err != nil {
		return "Error occured.", false
	}
	return "", true
}

func deleteSavedNotes(noteName string, from string) bool {
	f, err := ioutil.ReadFile("notes/data.json")
	var obj []Notes
	if err != nil {
		fmt.Println("Delete note : Failed to load note file, read file error")
		return false
	}
	err = json.Unmarshal(f, &obj)
	if err != nil {
		fmt.Println("Delete note : Failed to delete note, Unmarshal error")
		return false
	}
	isFound := false
	for index, data := range obj {
		if data.NoteName == noteName && data.From == from {
			obj = append(obj[:index], obj[index+1:]...)
			isFound = true
			break
		}
	}
	if !isFound {
		return false
	}
	dataJSON, err := json.Marshal(obj)
	if err != nil {
		fmt.Println("Failed to ADD ID to log, marshal error")
		return false
	}
	//fmt.Println(string(libJSON))
	if err := ioutil.WriteFile("notes/data.json", dataJSON, 0644); err != nil {
		return false
	}
	return true
}

func getSavedNotes(noteName string, from string) (string, bool) {
	f, err := ioutil.ReadFile("notes/data.json")
	var obj []Notes
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
		return "Error occured while getting notes.", false
	}
	err = json.Unmarshal(f, &obj)
	if err != nil {
		fmt.Println("Failed to get note, Unmarshal error")
		return "Error occured.", false
	}
	for _, name := range obj {
		if name.NoteName == noteName && name.From == from {
			return name.NoteText, true
		}
	}
	return "", false
}

func getNoteList(from string) string {
	f, err := ioutil.ReadFile("notes/data.json")
	var obj []Notes
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
		return "Error occured while getting notes."
	}
	err = json.Unmarshal(f, &obj)
	if err != nil {
		fmt.Println("Failed to get notes, Unmarshal error")
		return "Error occured."
	}
	noteList := "*Note List*"
	isFound := false
	if len(obj) > 0 {
		for _, name := range obj {
			if from == name.From {
				noteList += "\n - " + name.NoteName
				isFound = true
			}
		}
		if isFound {
			return noteList
		}
	}
	return noteList + " is Empty.\n why don't you try adding new notes.\nCommand : /save note-name message"
}

func checkImageData(fileName string) bool {
	file, err := os.Open("profile-images")
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
		return false
	}
	defer file.Close()

	list, _ := file.Readdirnames(0) // 0 to read all files and folders
	for _, name := range list {
		if name == fileName+".jpg" {
			return true
		}
	}
	return false
}

func getImageData(link string, fileName string) error {
	img, err := os.Create("profile-images/" + fileName + ".jpg")
	if err != nil {
		return errors.New("file not create")
	}
	defer img.Close()

	resp, _ := http.Get(link)
	if err != nil {
		return errors.New("response error")
	}
	defer resp.Body.Close()

	b, _ := io.Copy(img, resp.Body)
	fmt.Println("File size: ", b)
	return nil
}

func imageHandler(msg client.Message, newClient *client.WhatsappClient) {
	//fmt.Println(msg.ImageSource)
	newClient.SendText(msg.From, "Image not allowed.", msg.Info, true)
}

func stickerHandler(msg client.Message, newClient *client.WhatsappClient) {
	//fmt.Println(msg.StickerSource)
	//newClient.SendText(msg.From, "Sticker Banned.", msg.Info, true)
	stickerData, _ := os.Open("stickers/" + strconv.Itoa(rand.Intn(33)+1) + ".webp")
	newClient.SendSticker(msg.From, stickerData, msg.Info, true)
}

func actionHandler(msg client.Message, newClient *client.WhatsappClient) {
	fmt.Println(msg.Actions)
	if msg.Actions.Action == "add" {
		if partInterface, ok := msg.Actions.Participants["participants"].([]interface{}); ok {
			for _, v := range partInterface {
				userAdded := strings.Replace(v.(string), "@c.us", "@s.whatsapp.net", 1)
				sendMsg := "Welcome to Volcano Elite Group, " + newClient.GetName(userAdded) + ".\nEnjoy :)"
				newClient.SendText(msg.From, sendMsg, msg.Info, false)
			}
		}

	}
}

func getNumberValue(data string, operater string) (float64, float64, error) {
	splitData := strings.Split(data, operater)
	a1, err := strconv.ParseFloat(strings.TrimSpace(splitData[0]), 64)
	if err != nil {
		return 0, 0, errors.New("not found")
	}
	b1, err := strconv.ParseFloat(strings.TrimSpace(splitData[1]), 64)
	if err != nil {
		return 0, 0, errors.New("not found")
	}
	return a1, b1, nil
}

func handleCalcCommand(userData string, msg client.Message, newClient *client.WhatsappClient) {
	if userData == "" {
		newClient.SendText(msg.From, "Example : /calc 1+2. Supports : +,-,*,/", msg.Info, true)
	} else {
		if strings.Contains(userData, "+") {
			a, b, err := getNumberValue(userData, "+")
			if err == nil {
				result := a + b
				newClient.SendText(msg.From, userData+" = "+strconv.FormatFloat(result, 'f', 2, 64), msg.Info, true)
			} else {
				newClient.SendText(msg.From, "Error occured in operation", msg.Info, true)
			}
		} else if strings.Contains(userData, "-") {
			a, b, err := getNumberValue(userData, "-")
			if err == nil {
				result := a - b
				newClient.SendText(msg.From, userData+" = "+strconv.FormatFloat(result, 'f', 2, 64), msg.Info, true)
			} else {
				newClient.SendText(msg.From, "Error occured in operation", msg.Info, true)
			}
		} else if strings.Contains(userData, "*") {
			a, b, err := getNumberValue(userData, "*")
			if err == nil {
				result := a * b
				newClient.SendText(msg.From, userData+" = "+strconv.FormatFloat(result, 'f', 2, 64), msg.Info, true)
			} else {
				newClient.SendText(msg.From, "Error occured in operation", msg.Info, true)
			}
		} else if strings.Contains(userData, "/") {
			a, b, err := getNumberValue(userData, "/")
			if err == nil {
				if b == 0 {
					newClient.SendText(msg.From, userData+" = not defined", msg.Info, true)
				} else {
					result := a / b
					newClient.SendText(msg.From, userData+" = "+strconv.FormatFloat(result, 'f', 2, 64), msg.Info, true)
				}
			} else {
				newClient.SendText(msg.From, "Error occured in operation", msg.Info, true)
			}
		} else {
			newClient.SendText(msg.From, "This operation is not supported.", msg.Info, true)
		}
	}
}

func handleProfileCommand(userData string, msg client.Message, newClient *client.WhatsappClient) {
	if userData == "" {
		userData = msg.From
	} else if strings.HasPrefix(userData, "@") {
		userData = strings.Trim(userData, "@")
		userData = userData + "@s.whatsapp.net"
		fmt.Println(userData)
	} else {
		newClient.SendText(msg.From, "Command not used properly.", msg.Info, true)
		return
	}
	profileLink, profileTag := newClient.GetProfileImage(userData)
	if profileLink == "" {
		newClient.SendText(msg.From, "Profile Image not found", msg.Info, true)
	} else {
		if !checkImageData(profileTag) {
			err := getImageData(profileLink, profileTag)
			if err != nil {
				newClient.SendText(msg.From, "Error Occured while getting profile image.", msg.Info, true)
			}
		}
		imgData, _ := ioutil.ReadFile("profile-images/" + profileTag + ".jpg")
		newClient.SendImage(msg.From, "", "image/jpeg", imgData, msg.Info, true)
	}
}

func handleHiCommand(msg whatsapp.TextMessage, newClient *client.WhatsappClient) {
	var userData string
	if msg.Info.FromMe == true {
		newClient.SendText(msg.Info.RemoteJid, "Hello ME 😂", msg.Info, false)
	} else if msg.Info.Source.Participant != nil {
		userData = *msg.Info.Source.Participant
		//userData = strings.Trim(userData, "@s.whatsapp.net")
		newClient.SendText(msg.Info.RemoteJid, "Hello "+newClient.GetName(userData)+" !!\nNice to meet you 😊", msg.Info, true)
	} else {
		userData = msg.Info.RemoteJid
		//userData = strings.Trim(userData, "@s.whatsapp.net")
		newClient.SendText(msg.Info.RemoteJid, "Hello "+newClient.GetName(userData)+" !!\nNice to meet you 😊", msg.Info, true)
	}
}
