package app

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func GetFirebaseApp() *firebase.App {
	ctx := context.Background()
	opt := option.WithCredentialsFile("secrets/serviceKey.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatal(err)
	}

	return app
}

func GetFirestore() *firestore.Client {
	ctx := context.Background()
	app := GetFirebaseApp()

	db, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	return db
}
