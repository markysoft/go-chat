package handlers

import (
	"database/sql"
	"fmt"
	"go-star/common"
	"go-star/common/dal"
	"go-star/handlers/components"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"github.com/starfederation/datastar-go/datastar"
)

type Handlers struct {
	logger *slog.Logger
	db     *sql.DB
	nc     *nats.Conn
}

type ChatItem struct {
	Message  string `json:"message"`
	Username string `json:"username"`
}

var subject = "chat-messages"

func NewHandlers(logger *slog.Logger, db *sql.DB, nc *nats.Conn) *Handlers {
	return &Handlers{
		logger: logger,
		db:     db,
		nc:     nc,
	}
}
func (app *Handlers) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)
	app.logger.Error(err.Error(), "method", method, "uri", uri)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (h *Handlers) ListRooms() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomList, err := dal.ListRooms(h.db)
		if err != nil {
			h.serverError(w, r, fmt.Errorf("failed to list rooms: %w", err))
			return
		}

		templ.Handler(components.RoomsList(roomList)).ServeHTTP(w, r)
	}
}

func (h *Handlers) RoomPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomIdStr := chi.URLParam(r, "id")
		roomId, err := strconv.ParseInt(roomIdStr, 10, 64)
		if err != nil {
			h.serverError(w, r, fmt.Errorf("failed to parse room ID: %w", err))
			return
		}

		room, err := dal.GetRoom(h.db, roomId)
		if err != nil {
			h.serverError(w, r, fmt.Errorf("failed to get room: %w", err))
			return
		}

		chatter, err := h.getChatter(w, r)
		if err != nil {
			h.serverError(w, r, err)
			return
		}

		templ.Handler(components.RoomPage(*room, *chatter, components.RoomSignals{
			RoomId: room.ID,
			UserId: chatter.ID,
		})).ServeHTTP(w, r)
	}
}

func (h *Handlers) ListMessages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		roomSignals := &components.RoomSignals{}
		if err := datastar.ReadSignals(r, roomSignals); err != nil {
			h.serverError(w, r, fmt.Errorf("failed to read room signals: %w", err))
			return
		}

		userID, err := common.GetUserID(w, r)
		if err != nil {
			h.serverError(w, r, fmt.Errorf("failed to get or generate user ID: %w", err))
			return
		}

		log.Printf("Client connected to messages stream with userID: %s", userID)
		sse := datastar.NewSSE(w, r)
		patchMessages(h, sse, userID, roomSignals.RoomId)
		// Create a channel to receive messages from NATS
		messageChan := make(chan string, 10)
		// Subscribe to NATS and forward messages to the channel
		sub, err := h.nc.Subscribe(subject, func(msg *nats.Msg) {
			log.Println("message received from NATS")
			data := string(msg.Data)
			select {
			case messageChan <- data:
			default:
				// Channel is full, drop the message
				log.Printf("Message channel full, dropping message: %s", data)
			}
		})
		if err != nil {
			h.serverError(w, r, fmt.Errorf("failed to subscribe to messages: %w", err))
			return
		}
		defer sub.Unsubscribe()

		for {
			select {
			case <-r.Context().Done():
				log.Println("Client disconnected from messages stream")
				return
			case message := <-messageChan:
				log.Printf("Sending message to client: %s", message)
				if message == "" {
					continue
				}
				ctrl := patchMessages(h, sse, userID, roomSignals.RoomId)
				switch ctrl {
				case 1:
					continue
				}
			}
		}
	}
}

func (h *Handlers) SendMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		message := &ChatItem{}
		if err := datastar.ReadSignals(r, message); err != nil {
			h.serverError(w, r, fmt.Errorf("failed to read message signal: %w", err))
			return
		}

		chatter, err := h.getChatter(w, r)
		if err != nil {
			h.serverError(w, r, err)
			return
		}
		
		message.Username = chatter.Username

		_, err = dal.InsertMessage(h.db, chatter.ID, 1, message.Message)
		if err != nil {
			h.serverError(w, r, fmt.Errorf("failed to insert message: %w", err))
			return
		}

		formattedMessage := fmt.Sprintf("%s:%s", message.Username, message.Message)
		err = h.nc.Publish(subject, []byte(formattedMessage))
		if err != nil {
			h.serverError(w, r, fmt.Errorf("failed to publish message: %w", err))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func patchMessages(h *Handlers, sse *datastar.ServerSentEventGenerator, username string, roomId int64) int {
	allMessages, err := dal.ListMessagesForRoom(h.db, roomId)
	if err != nil {
		log.Printf("Failed to list messages: %v", err)
		return 1
	}
	err = sse.PatchElementTempl(components.Messages(allMessages, username))
	if err != nil {
		log.Printf("Failed to send message to client: %v", err)
		return 0
	}
	return 0
}

func (app *Handlers) getChatter(w http.ResponseWriter, r *http.Request) (*dal.Chatter, error) {
	userID, err := common.GetUserID(w, r)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("failed to initialize user session: %w", err))
		return nil, err
	}

	chatter, _ := dal.GetChatterByUsername(app.db, userID)
	if chatter == nil {
		totalChatters, err := dal.TotalChatters(app.db)
		if err != nil {
			app.serverError(w, r, fmt.Errorf("failed to get total chatters: %w", err))
			return nil, err
		}
		chatter, err = dal.InsertChatter(app.db, userID, fmt.Sprintf("User No. %d", totalChatters+1))
		if err != nil {
			app.serverError(w, r, fmt.Errorf("failed to create new chatter: %w", err))
			return nil, err
		}
	}
	return chatter, nil
}
