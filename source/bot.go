package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type (
	Items struct {
		Item []Item `json:"items"`
	}

	Item struct {
		Type       string `json:"type,omitempty"`
		Imageurl   string `json:"imageurl,omitempty"`
		Muatienmat string `json:"muatienmat,omitempty"`
		Muack      string `json:"muack,omitempty"`
		Bantienmat string `json:"bantienmat,omitempty"`
		Banck      string `json:"banck,omitempty"`
	}
)

const (
	GetStartedFB = "GetStarted"
	RateFB       = "rate"
)

var (
	ItemList     *Items
	ItemGroupMap = make(map[string]int)
)

func processMessage(event *Messaging) {
	sendAction(event.Sender, MarkSeen)
	sendAction(event.Sender, TypingOn)

	if event.Message.QuickReply != nil {
		processQuickReply(event)
		return
	}

	text := strings.ToLower(strings.TrimSpace(event.Message.Text))

	if text == "rate" {
		ItemGroupMap[event.Sender.ID] = 1
		sendItemList(event.Sender)
	} else {
		sendText(event.Sender, strings.ToUpper(event.Message.Text))
	}

	sendAction(event.Sender, TypingOff)
}

func processQuickReply(event *Messaging) {
	recipient := event.Sender
	ItemGroup := ItemGroupMap[event.Sender.ID]

	switch event.Message.QuickReply.Payload {
	case "Next":
		var i int
		if ItemGroup*10 >= len(ItemList.Item) {
			ItemGroup = 1
		} else {
			ItemGroup++
		}

		ItemGroupMap[event.Sender.ID] = ItemGroup
		quickRep := []QuickReply{}

		for i = 10 * (ItemGroup - 1); i < 10*ItemGroup && i < len(ItemList.Item); i++ {
			item := ItemList.Item[i]
			quickRep = append(quickRep,
				QuickReply{
					ContentType: "text",
					Title:       item.Type,
					ImageUrl:    item.Imageurl,
					Payload:     item.Type,
				},
			)
		}

		quickRep = append(quickRep,
			QuickReply{
				ContentType: "text",
				Title:       "xem tiếp",
				Payload:     "Next",
			},
		)

		sendTextWithQuickReply(recipient, "Bot ngu si cung cấp chức năng xem tỉ giá giữa các ngoại tệ và đồng Việt Nam. Được cập nhật hàng ngày từ dữ liệu của ngân hàng Đông Á.\n Mời bạn chọn ngoại tệ: ", quickRep)
	default:
		var Item Item

		for i := 10 * (ItemGroup - 1); i < 10*ItemGroup && i < len(ItemList.Item); i++ {
			if ItemList.Item[i].Type == event.Message.QuickReply.Payload {
				Item = ItemList.Item[i]
				break
			}
		}

		if len(Item.Type) == 0 {
			sendText(recipient, "Không tìm thấy thông tin về ngoại tệ này")
			return
		}

		sendText(recipient, fmt.Sprintf("Giá mua tiền mặt: %sđ\nGiá mua chuyển khoản: %sđ\nGiá bán tiền mặt: %sđ\nGía bán chuyển khoản: %sđ\nCảm ơn bạn đã sử dụng dịch vụ!", Item.Muatienmat, Item.Muack, Item.Bantienmat, Item.Banck))
	}
}

func sendItemList(recipient *User) {
	var (
		ok        bool
		i         int
		ItemGroup = ItemGroupMap[recipient.ID]
	)

	ItemList, ok = getItemDongA()

	if !ok {
		sendText(recipient, "Có lỗi trong quá trình xử lý. Bạn vui lòng thử lại sau bằng cách gửi 'rate' cho tôi nhé. Xin cảm ơn!")
		return
	}

	quickRep := []QuickReply{}
	for i = 10 * (ItemGroup - 1); i < 10*ItemGroup && i < len(ItemList.Item); i++ {
		item := ItemList.Item[i]
		quickRep = append(quickRep,
			QuickReply{
				ContentType: "text",
				Title:       item.Type,
				ImageUrl:    item.Imageurl,
				Payload:     item.Type,
			},
		)
	}

	quickRep = append(quickRep,
		QuickReply{
			ContentType: "text",
			Title:       "Xem tiếp",
			Payload:     "Next",
		},
	)

	sendTextWithQuickReply(recipient, "Bot ngu si cung cấp chức năng xem tỉ giá giữa các ngoại tệ và đồng Việt Nam. Được cập nhật hàng ngày từ dữ liệu của ngân hàng Đông Á\nMời bạn chọn ngoại tệ: ", quickRep)
}

func getItemDongA() (*Items, bool) {
	var items Items

	req, err := http.NewRequest("GET", "https://www.dongabank.com.vn/exchange/export", nil)
	if err != nil {
		log.Println("getItemDongA: NewRequest ", err.Error())
		return &items, false
	}

	client := &http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("getItemDongA: client.Do ", err.Error())
		return &items, false
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	body = body[1 : len(body)-1]

	err = json.Unmarshal([]byte(body), &items)

	if err != nil {
		log.Println("getItemDongA: json.NewDecoder: ", err.Error())
		return &items, false
	}

	return &items, true
}

func processPostBack(event *Messaging) {
	sendAction(event.Sender, MarkSeen)
	sendAction(event.Sender, TypingOn)

	switch event.PostBack.Payload {
	case GetStartedFB, RateFB:
		ItemGroupMap[event.Sender.ID] = 1
		sendItemList(event.Sender)
	}
	sendAction(event.Sender, TypingOff)
}
