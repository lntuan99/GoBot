package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const (
	FBMessageURL    = "https://graph.facebook.com/v6.0/me/messages"
	PageToken       = "EAAISUVhDWSYBAF6HJGYhnuMU2fTR0dSSH5YMToqPkSYY0nOjjvtT84SRc0jV2Rc7TtmxGvGKJtpvhyWZCZCCLVaZCLVkZAfADxAZBagcRekf6Cba7L7QPFRSIiWI8C6tvsRixisrvfOdyYdQw345ZAF9FG7TEezBBELvtJZBEmJq6b9zTNsCPZBe"
	MessageResponse = "RESPONSE"
	TypingOn        = "typing_on"
	TypingOff       = "typing_off"
	MarkSeen        = "mark_seen"
)

type (
	Request struct {
		Object string `json:"object,omitempty"`
		Entry  []struct {
			ID        string      `json:"id,omitempty"`
			Time      int64       `json:"time,omitempty"`
			Messaging []Messaging `json:"messaging, omitempty"`
		} `json:"entry,omitempty"`
	}

	Messaging struct {
		Sender    *User     `json:"sender,omitempty"`
		Recipient *User     `json:"recipient,omitempty"`
		Timestamp int       `json:"timestamp,omitempty"`
		Message   *Message  `json:"message,omitempty"`
		PostBack  *PostBack `json:"postback,omitempty"`
	}

	User struct {
		ID string `json:"id,omitempty"`
	}

	Message struct {
		MID        string      `json:"mid,omitempty"`
		Text       string      `json:"text,omitempty"`
		QuickReply *QuickReply `json:"quick_reply,omitempty"`
	}

	QuickReply struct {
		ContentType string `json:"content_type,omitempty"`
		Title       string `json:"title,omitempty"`
		ImageUrl    string `json:"image_url,omitempty"`
		Payload     string `json:"payload"`
	}

	PostBack struct {
		Title   string `json:"title,omitempty"`
		Payload string `json:"payload"`
	}

	ResponseMessage struct {
		MessageType string      `json:"messaging_type"`
		Recipient   *User       `json:"recipient"`
		Message     *ResMessage `json:"message,omitempty""`
		Action      string      `json:"sender_action,omitempty"`
	}

	ResMessage struct {
		Text       string       `json:"text,omitempty"`
		QuickReply []QuickReply `json:"quick_replies,omitempty"`
	}
)

type (
	PageProfile struct {
		Greeting       []Greeting       `json:"greeting,omitempty"`
		GetStarted     *GetStarted      `json:"get_started,omitempty"`
		PersistentMenu []PersistentMenu `json:"persistent_menu,omitempty"`
	}

	Greeting struct {
		Locale string `json:"locale,omitempty"`
		Text   string `json:"text,omitempty"`
	}

	GetStarted struct {
		Payload string `json:"payload,omitempty"`
	}

	PersistentMenu struct {
		Locale   string `json:"locale"`
		Composer bool   `json:"composer_input_disabled"`
		CTAs     []CTA  `json:"call_to_actions"`
	}

	CTA struct {
		Type    string `json:"type"`
		Title   string `json:"title"`
		URL     string `json:"url,omitempty"`
		Payload string `json:"payload"`
		CTAs    []CTA  `json:"call_to_actions,omitempty"`
	}
)

func sendFBRequest(url string, m interface{}) error {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&m)

	if err != nil {
		log.Println("sendFBRequest.json.NewEncoder: " + err.Error())
		return err
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Println("sendFBRequest.json.NewEncoder: " + err.Error())
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.URL.RawQuery = "access_token=" + PageToken
	client := &http.Client{Timeout: time.Second * 30}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("sendFBRequest.client.Do: " + err.Error())
		return err
	}

	defer resp.Body.Close()

	return nil
}

func sendTextWithQuickReply(recipient *User, message string, replies []QuickReply) error {
	m := ResponseMessage{
		MessageType: MessageResponse,
		Recipient:   recipient,
		Message: &ResMessage{
			Text:       message,
			QuickReply: replies,
		},
	}

	return sendFBRequest(FBMessageURL, &m)
}

func sendText(recipient *User, message string) error {
	return sendTextWithQuickReply(recipient, message, nil)
}

func sendAction(recipient *User, action string) error {
	m := ResponseMessage{
		MessageType: MessageResponse,
		Recipient:   recipient,
		Action:      action,
	}

	return sendFBRequest(FBMessageURL, &m)
}

func registerGreetingMenu() bool {
	profile := PageProfile{
		Greeting: []Greeting{
			{
				Locale: "default",
				Text:   "Dịch vụ cung cấp thông tin tỉ giá hối đoái",
			},
		},

		GetStarted: &GetStarted{Payload: GetStartedFB},

		PersistentMenu: []PersistentMenu{
			{
				Locale:   "default",
				Composer: false,
				CTAs: []CTA{
					{
						Type:    "postback",
						Title:   "Tỉ giá hối đoái",
						Payload: RateFB,
					},
				},
			},
		},
	}

	err := sendFBRequest("https://graph.facebook.com/v6.0/me/messenger_profile", &profile)

	if err != nil {
		log.Println("registerGreetingMenu: ", err.Error())
		return false
	}

	return true
}
