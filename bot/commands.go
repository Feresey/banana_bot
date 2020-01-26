package bot

import (
	"encoding/json"
	"fmt"

	"github.com/Feresey/banana_bot/db"
	"github.com/Feresey/banana_bot/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func  processMessage(msg *model.Message) {
	// предполагая что у меня руки из жопы я оставлю это
	defer func() {
		if err := recover(); err != nil {
			log.Error("Fall in panic", err)
		}
	}()

	if debug {
		data, _ := json.MarshalIndent(msg, "", "\t")
		log.Info(string(data))
	} else {
		log.Infof("[%s] %s", msg.From.UserName, msg.Text)
	}

	switch msg.Chat.Type {
	case "private":
		if err := privateMessage(msg); err != nil {
			log.Warn(err)
		}
	case "group", "supergroup":
		if err := groupMessage(msg); err != nil {
			log.Warn(err)
		}
	}
}

func  groupMessage(msg *model.Message) error {
	cmd := msg.Command()
	if cmd == "" {
		return nil
	}

	isPublic := isPublicMethod(cmd)
	isAdmin := isAdmin(msg)

	if !isPublic {
		if !isAdmin {
			reply := model.NewReply(msg)
			reply.Text = "Только админам можно"
			sendMsg(reply)
			return nil
		}
		return processAdminActions(msg)
	}
	return processPublicActions(msg)
}

func  processPublicActions(msg *model.Message) error {
	var (
		cmd   = msg.Command()
		reply *model.Reply
		err   error
	)

	switch cmd {
	case "report":
		reply, err = report(msg)
	}
	if err != nil {
		return err
	}

	if reply != nil {
		sendMsg(reply)
	}
	return nil
}

func  processAdminActions(msg *model.Message) error {
	var (
		cmd   = msg.Command()
		reply *model.Reply
		err   error
	)

	switch cmd {
	case "report":
		reply, err = report(msg)
	case "ban":
		reply, err = ban(msg)
	case "warn":
		reply, err = warn(msg, true)
	case "unwarn":
		reply, err = warn(msg, false)
	}
	if err != nil {
		return err
	}

	if reply != nil {
		sendMsg(reply)
	}
	return nil
}

func  report(msg *model.Message) (*model.Reply, error) {
	subscribed, err := db.Report(msg.Chat.ID)
	if err != nil {
		return nil, err
	}

	for _, val := range subscribed {
		reply := model.Reply{
			MessageConfig: &tgbotapi.MessageConfig{
				BaseChat: tgbotapi.BaseChat{
					ChatID: int64(val),
				},
			},
		}
		sendMsg(reply)
	}

	reply := model.NewReply(msg)
	reply.Text = "Админы призваны!"
	return reply, nil
}

func  ban(msg *model.Message) (*model.Reply, error) {
	reply := model.NewReply(msg)
	if msg.ReplyToMessage == nil {
		reply.Text = "Надо использовать команду ответом на сообщение"
		return reply, nil
	}
	user := msg.ReplyToMessage.From

	person := &model.Person{
		ChatID: msg.Chat.ID,
		UserID: user.ID,
	}

	if r := protect(person, msg.From.ID); r != nil {
		return r, nil
	}

	err := kickMember(person)
	reply.Text = "F"
	return reply, err
}

func  warn(msg *model.Message, add bool) (*model.Reply, error) {
	reply := model.NewReply(msg)

	if msg.ReplyToMessage == nil {
		if add {
			reply.Text = "Не указано кому /warn кидать"
		} else {
			reply.Text = "Не указано кому /unwarn кидать"
		}
		return reply, nil
	}

	user := msg.ReplyToMessage.From
	person := &model.Person{
		ChatID: msg.Chat.ID,
		UserID: user.ID,
	}

	if r := protect(person, msg.From.ID); r != nil {
		return r, nil
	}

	total, err := db.Warn(person, add)
	if err != nil {
		return nil, err
	}

	switch {
	case total < maxWarn:
		reply.Text = fmt.Sprintf("@%s, Предупреждение %d/%d", user.UserName, total, maxWarn)
	case total == maxWarn:
		reply.Text = fmt.Sprintf("@%s, Последнее предупреждение!", user.UserName)
	default:
		reply.Text = "F"
	}

	if total > maxWarn {
		err = kickMember(person)
	}
	return reply, err
}

func  privateMessage(msg *model.Message) error {
	cmd := msg.Command()
	// GetChat(tgbotapi.ChatConfig{ChatID: msg.Chat.ID})
	switch cmd {
	case "start":
		reply := model.NewReply(msg)
		reply.Text = "Приветствую, кожаный мешок"
		sendMsg(reply)
	default:
	}
	return nil
}

func protect(p *model.Person, id int) *model.Reply {
	reply := &model.Reply{
		MessageConfig: &tgbotapi.MessageConfig{
			BaseChat: tgbotapi.BaseChat{
				ChatID: p.ChatID,
			},
		},
	}

	switch p.UserID {
	case 425496698:
		reply.Text = "Я не могу пойти против создателя. Ave Banana!"
	case 1066353768:
		reply.Text = "Бан бану рознь."
	case id:
		reply.Text = "Мазохизм не приветствуется."
	default:
		return nil
	}
	return reply
}

func  kickMember(p *model.Person) error {
	kick := &tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: p.ChatID,
			UserID: p.UserID,
		},
	}

	resp, err := bot.KickChatMember(*kick)
	if err != nil {
		log.Warnf("%#v : [%s]", resp, err.Error())
	} else {
		log.Infof("Succeccfully kicked: ID:[%d]", kick.UserID)
	}
	return nil
}
