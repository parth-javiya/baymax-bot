package whatsappclient

import (
	"encoding/gob"
	"fmt"
	"os"
	"os/signal"
	"time"

	whatsapp "github.com/Rhymen/go-whatsapp"
	qrcode "github.com/skip2/go-qrcode"
)

// WhatsappClient connect
type WhatsappClient struct {
	wac *whatsapp.Conn
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

	fmt.Println("Whatsapp Connected!")
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

// SendText send text message
func (wp *WhatsappClient) SendText(to string, text string) {
	reply := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: to,
		},
		Text: text,
	}

	msgID, err := wp.wac.Send(reply)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending message: %v", err)
	} else {
		fmt.Println("Message Sent -> ID : " + msgID)
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
