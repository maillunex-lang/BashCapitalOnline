package sim

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"

	"bashcapital/server/internal/model"
)

func (e *Engine) PostChat(firmID, channel, text string) (*model.ChatMessage, error) {
	text = strings.TrimSpace(text)
	if text == "" || len([]rune(text)) > 280 {
		return nil, fmt.Errorf("сообщение 1…280 символов")
	}
	if channel == "" {
		channel = "global"
	}
	author := "Игрок"
	isAI := false
	if firmID != "" {
		if f := e.World.Firms[firmID]; f != nil {
			author = f.Name
			isAI = f.IsAI
			f.LastChatTick = e.World.Tick
		}
	}
	msg := model.ChatMessage{
		ID: uuid.NewString(), Tick: e.World.Tick, FirmID: firmID, Author: author,
		Channel: channel, Text: text, IsAI: isAI, CreatedAt: time.Now().Unix(),
	}
	e.World.Chat = append(e.World.Chat, msg)
	e.trimChat()
	return &msg, nil
}

func (e *Engine) trimChat() {
	max := e.Catalog.Economy.Chat.MaxMessages
	if max <= 0 {
		max = 120
	}
	if len(e.World.Chat) > max {
		e.World.Chat = e.World.Chat[len(e.World.Chat)-max:]
	}
}

func (e *Engine) runAIChat(log *[]string) {
	cfg := e.Catalog.Economy.Chat
	capN := cfg.AIMessagesPerTickCap
	if capN <= 0 {
		capN = 2
	}
	minGap := cfg.MinTicksBetweenAIMsg
	if minGap <= 0 {
		minGap = 2
	}
	if len(e.World.LastTickLog) == 0 && rand.Float64() > 0.25 {
		return
	}
	posted := 0
	for _, firm := range e.World.Firms {
		if !firm.IsAI || posted >= capN {
			continue
		}
		if e.World.Tick-firm.LastChatTick < minGap {
			continue
		}
		persona := e.Catalog.PersonaByID[firm.PersonaID]
		if persona.ID == "" {
			persona.ChatTone = "friendly_local"
			persona.Skill = "solid"
		}
		// probability by aggression/skill
		chance := 0.18 + persona.Aggression*0.15
		if persona.Skill == "noob" {
			chance += 0.12
		}
		if rand.Float64() > chance {
			continue
		}
		text := e.composeAIChat(firm, persona)
		if text == "" {
			continue
		}
		_, _ = e.PostChat(firm.ID, "global", text)
		posted++
		*log = append(*log, fmt.Sprintf("Чат: %s — %s", firm.Name, text))
	}
}

func (e *Engine) composeAIChat(firm *model.Firm, p model.AIPersona) string {
	hooks := e.World.LastTickLog
	eventHint := ""
	for i := len(hooks) - 1; i >= 0 && i >= len(hooks)-4; i-- {
		h := hooks[i]
		if strings.Contains(h, "Розница") || strings.Contains(h, "Биржа") || strings.Contains(h, "Залежь") || strings.Contains(h, "Событие") || strings.Contains(h, "дорог") {
			eventHint = h
			break
		}
	}
	switch p.ChatTone {
	case "excited_noob":
		if eventHint != "" {
			return pick([]string{
				fmt.Sprintf("Ребят, это нормально? У меня в ленте: «%s» 😅", short(eventHint, 60)),
				"Только открыл лавку и уже боюсь цен. Городское молоко хоть и хилое, но дешёвое!",
				"Кто-нибудь подскажет, почему бутылки кончаются быстрее молока?",
			})
		}
		return pick([]string{
			"Я вроде всё правильно сделал… или это так всегда у новичков?",
			"Уфа красивая, а экономика кусается. Пойду ещё раз посмотрю залежи.",
		})
	case "dry_humor":
		if strings.Contains(eventHint, "Залежь") {
			return "Залежь кончилась — классика. Ресурс как парковка: пока искал, уже в другом районе."
		}
		if p.Style == "wholesale" {
			return pick([]string{
				"Розница пусть играет в витрины. Мне бы стабильный опт и чтобы фуры не стояли.",
				"Кто ставит хлеб по миллиону — тот не конкурент, тот баг. Коридор цен спасёт нервы.",
			})
		}
		return "Мебель без ткани и ДСП — это не бизнес, это молитва."
	case "gruff":
		return pick([]string{
			"Руда есть — разговаривать меньше, плавить больше.",
			"Белорецк не для нежных цен. Кто на городском песке выезжает — сам виноват: качество никудышное.",
		})
	case "market_nerd":
		if strings.Contains(eventHint, "Биржа") {
			return fmt.Sprintf("По спреду сегодня интересно. Зацепил: %s", short(eventHint, 70))
		}
		return pick([]string{
			"Смотрю арбитраж Уфа↔Белорецк. Три часа дороги — уже в цене надо сидеть.",
			"Городской опт с quality 1.x годится только чтобы не встать. Дальше — свои цепочки.",
		})
	default: // friendly_local
		if strings.Contains(eventHint, "Событие") {
			return "Администрация снова шевелится. Надо не проспать субсидии на молочку."
		}
		if strings.Contains(eventHint, "Розница") {
			return "В районе магазинчик ожил. Главное — не ставить цену как для космонавтов."
		}
		return pick([]string{
			"Доброе утро, республика. Кто на вишнёво-молочный заход делает ставки?",
			fmt.Sprintf("У %s касса пока дышит. Кто на пару по бутылкам?", firm.Name),
		})
	}
}

func pick(opts []string) string {
	if len(opts) == 0 {
		return ""
	}
	return opts[rand.Intn(len(opts))]
}

func short(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
