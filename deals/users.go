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
}

type FirestoreUser struct {
	DiscordUserID string
	PostalCode    string
	Merchants     []int
	Items         []string
}

func UpdateUser(firestoreClient *firestore.Client, userMessage UserMessage) {
	ref := firestoreClient.Collection("users").Doc(userMessage.DiscordUserID)
	doc, err := ref.Get(context.Background())

	if status.Code(err) == codes.NotFound {
		firestoreUser := FirestoreUser{
			DiscordUserID: userMessage.DiscordUserID,
			PostalCode:    userMessage.PostalCode,
		}

		if userMessage.Merchant == 0 && userMessage.Item == "" {
			log.Infof("Empty usermessage %v", userMessage)
			return
		}

		firestoreUser.Items = []string{userMessage.Item}
		firestoreUser.Merchants = []int{userMessage.Merchant}

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

		var updates []firestore.Update

		if userMessage.Merchant != 0 {
			updates = append(updates, firestore.Update{Path: "Merchants", Value: appendMerchant(firestoreUser.Merchants, userMessage.Merchant)})
		}

		if userMessage.Item != "" {
			updates = append(updates, firestore.Update{Path: "Items", Value: appendItem(firestoreUser.Items, userMessage.Item)})
		}

		_, err = ref.Update(context.Background(), updates)
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

func appendMerchant(merchants []int, merchant int) []int {

	for _, m := range merchants {
		if merchant == m {
			return merchants
		}
	}
	return append(merchants, merchant)
}

func appendItem(items []string, item string) []string {

	for _, i := range items {
		if item == i {
			return items
		}
	}
	return append(items, item)
}
