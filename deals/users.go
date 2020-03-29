package deals

import (
	"cloud.google.com/go/firestore"
	"context"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserMessage struct {
	DiscordUserID string `json:"discord_user_id"`
	PostalCode    string `json:"postal_code"`
	Item          string `json:"item"`
	Merchant      int    `json:"merchant"`
	Remove        bool   `json:"remove"`
}

type FirestoreUser struct {
	DiscordUserID string
	PostalCode    string
	Items         map[string][]int
}

func UpdateUser(firestoreClient *firestore.Client, userMessage UserMessage) {
	ref := firestoreClient.Collection("users").Doc(userMessage.DiscordUserID)
	doc, err := ref.Get(context.Background())

	if status.Code(err) == codes.NotFound {
		firestoreUser := FirestoreUser{
			DiscordUserID: userMessage.DiscordUserID,
			PostalCode:    userMessage.PostalCode,
		}
		if userMessage.Item != "" && userMessage.Merchant != 0 {
			firestoreUser.Items = map[string][]int{userMessage.Item: {userMessage.Merchant}}
		}
		_, err := ref.Create(context.Background(), firestoreUser)
		if err != nil {
			log.Error("Error creating doc", firestoreUser, err)
			return
		}
	} else if status.Code(err) == codes.OK {
		var firestoreUser FirestoreUser
		err := doc.DataTo(&firestoreUser)
		if err != nil {
			log.Error("Error parsing doc", doc, err)
			return
		}
		if firestoreUser.Items == nil {
			firestoreUser.Items = make(map[string][]int)
		}
		merchants := firestoreUser.Items[userMessage.Item]
		firestoreUser.Items[userMessage.Item] = appendUnique(merchants, userMessage.Merchant)
		_, err = ref.Update(context.Background(), []firestore.Update{{Path: "items", Value: firestoreUser.Items}})
		if err != nil {
			log.Error("Error updating doc", firestoreUser.Items, err)
			return
		}
	} else {
		log.Error("Firestore error", err)
		return
	}
	log.Info("Successfully processed message", userMessage)
}

func appendUnique(guys []int, guy int) []int {

	for _, g := range guys {
		if guy == g {
			return guys
		}
	}
	return append(guys, guy)
}
