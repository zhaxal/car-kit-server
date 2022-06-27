package app

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/filipkroca/teltonikaparser"
	"github.com/thedevsaddam/gojsonq"
	"log"
)

type Trip struct {
	StartTime int
	EndTime   int
	Status    string
}

type Element struct {
	Name  string
	Value interface{}
	Units interface{}
}

type Avl struct {
	Altitude int
	Lat      int
	Lng      int
	Speed    int
	Time     int
	Elements []Element
}

type Device struct {
	IMEI string
}

func isDuplicate(documents interface{}, time int) bool {
	packets, _ := documents.([]interface{})
	for _, val := range packets {
		if time == int(val.(int64)) {
			return true
		}
	}
	return false
}

func AddDevice(device teltonikaparser.Decoded, ctx context.Context, client *firestore.Client) error {
	var previousPackets []int
	humanDecoder := teltonikaparser.HumanDecoder{}

	dsnap, err := client.Collection("devices").Doc(device.IMEI).Get(ctx)
	deviceDoc := dsnap.Data()

	if err != nil {
		log.Printf("An error has occurred: %s", err)
	}

	for _, val := range device.Data {

		previousPackets = append(previousPackets, int(val.Utime))

		if deviceDoc != nil {
			if isDuplicate(deviceDoc["PreviousPackets"], int(val.Utime)) {
				continue
			}
		}

		
		

		log.Printf("seconds %v", val.Utime)

		tripStatus := "none"

		packet := Avl{
			Altitude: int(val.Altitude),
			Lat:      int(val.Lat),
			Lng:      int(val.Lng),
			Speed:    int(val.Speed),
			Time:     int(val.Utime),
		}

		for _, ioel := range val.Elements {

			decoded, err := humanDecoder.Human(&ioel, "FMBXY")
			if err != nil {
				log.Printf("Error when converting human, %v\n", err)
			}

			if val, err := (*decoded).GetFinalValue(); err != nil {
				log.Printf("Unable to GetFinalValue() %v", err)
			} else if val != nil {
				units := gojsonq.New().File("secrets/units.json")
				res := units.Where("Property Name", "=", decoded.AvlEncodeKey.PropertyName).First()
				converted, _ := res.(map[string]interface{})

				element := Element{
					Name:  decoded.AvlEncodeKey.PropertyName,
					Value: val,
					Units: converted["Units"],
				}

				if element.Name == "Trip" {

					var active uint8 = 1
					var inactive uint8 = 0

					if element.Value == active {
						tripStatus = "active"

					} else if element.Value == inactive {
						tripStatus = "inactive"

					}
				}

				packet.Elements = append(packet.Elements, element)
			}
		}

		iter := client.Collection("devices").Doc(device.IMEI).Collection("trips").Where("Status", "==", "active").Limit(1).Documents(ctx)

		docs, err := iter.GetAll()
		if err != nil {
			log.Printf("Unable to GetAll() %v", err)
		}

		switch tripStatus {
		case "active":
			var trip Trip

			if len(docs) > 0 {
				doc := docs[0]
				doc.DataTo(&trip)

				_, _, err = doc.Ref.Collection("packets").Add(ctx, packet)

				if err != nil {
					log.Printf("An error has occurred: %s", err)
				}

			} else {
				trip = Trip{
					StartTime: packet.Time,
					EndTime:   0,
					Status:    "active",
				}

				tripRef, _, err := client.Collection("devices").Doc(device.IMEI).Collection("trips").Add(ctx, trip)

				if err != nil {
					log.Printf("An error has occurred: %s", err)
				}

				_, _, err = tripRef.Collection("packets").Add(ctx, packet)

				if err != nil {
					log.Printf("An error has occurred: %s", err)
				}
			}

		case "inactive":
			var trip Trip
			if len(docs) > 0 {
				doc := docs[0]
				doc.DataTo(&trip)

				trip.EndTime = packet.Time
				trip.Status = "inactive"

				_, err = doc.Ref.Set(ctx, trip)

				fmt.Println("ended trip")

				if err != nil {
					log.Printf("An error has occurred: %s", err)
				}
			}
		}

		_, _, err = client.Collection("devices").Doc(device.IMEI).Collection("packets").Add(ctx, packet)

		if err != nil {
			log.Printf("An error has occurred: %s", err)
		}
	}

	lastPacket := device.Data[len(device.Data)-1]

	_, err = client.Collection("devices").Doc(device.IMEI).Set(ctx, map[string]interface{}{
		"IMEI":            device.IMEI,
		"Lat":             lastPacket.Lat,
		"Lng":             lastPacket.Lng,
		"PreviousPackets": previousPackets,
	}, firestore.MergeAll)

	if err != nil {
		log.Printf("An error has occurred: %s", err)
	}

	return err
}
