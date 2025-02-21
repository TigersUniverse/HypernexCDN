package cdn

import (
	"HypernexCDN/api"
	"HypernexCDN/api/search"
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"strings"
)

var mongoclient *mongo.Client
var uploadsCollection *mongo.Collection
var avatarsCollection *mongo.Collection
var worldsCollection *mongo.Collection
var avatarPopularityCollection *mongo.Collection
var worldPopularityCollection *mongo.Collection

func ConnectToMongo(uri string) {
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(nil)
	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}
	mongoclient = client
	var result bson.M
	if err := mongoclient.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
		panic(err)
	}
	fmt.Println("Connected to MongoDB!")
	uploadsCollection = mongoclient.Database("main").Collection("uploads")
	avatarsCollection = mongoclient.Database("main").Collection("avatars")
	worldsCollection = mongoclient.Database("main").Collection("worlds")
	avatarPopularityCollection = mongoclient.Database("main").Collection("avatar_popularity")
	worldPopularityCollection = mongoclient.Database("main").Collection("world_popularity")
}

func GetFileMetaById(userid string, fileid string) *api.FileUpload {
	filter := bson.M{
		"UserId": userid,
		"Uploads": bson.M{
			"$elemMatch": bson.M{"FileId": fileid},
		},
	}
	var result api.UserUploads
	err := uploadsCollection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			fmt.Println("File not found for user")
			return nil
		} else {
			panic(err)
		}
	}
	if len(result.Uploads) > 0 {
		for _, upload := range result.Uploads {
			if upload.FileId == fileid {
				return &upload
			}
		}
	}
	return nil
}

func GetAvatarMetaFromFileId(userid string, fileid string) *search.AvatarMeta {
	filter := bson.M{
		"OwnerId": userid,
		"Builds": bson.M{
			"$elemMatch": bson.M{"FileId": fileid},
		},
	}
	var result search.AvatarMeta
	err := avatarsCollection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			fmt.Println("File not found for user")
			return nil
		} else {
			panic(err)
		}
	}
	return &result
}

func GetWorldMetaFromFileId(userid string, fileid string) *search.WorldMeta {
	filter := bson.M{
		"OwnerId": userid,
		"Builds": bson.M{
			"$elemMatch": bson.M{"FileId": fileid},
		},
	}
	var result search.WorldMeta
	err := worldsCollection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			fmt.Println("File not found for user")
			return nil
		} else {
			panic(err)
		}
	}
	return &result
}

func GetUploadData(userid string) *api.UserUploads {
	filter := bson.M{"UserId": userid}
	var result api.UserUploads
	err := uploadsCollection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			fmt.Println("User uploads not found!")
			return nil
		} else {
			panic(err)
		}
	}
	return &result
}

func UpdateUploadData(data *api.UserUploads) {
	filter := bson.M{"UserId": data.UserId}
	var result api.UserUploads
	err := uploadsCollection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			fmt.Println("User uploads not found!")
		} else {
			panic(err)
		}
	} else {
		update := bson.D{{"$set", data}}
		_, err2 := uploadsCollection.UpdateOne(context.TODO(), filter, update)
		if err2 != nil {
			panic(err2)
		}
	}
}

func GetOrCreatePopularity(id string) *api.PopularityObject {
	var collection *mongo.Collection
	split := strings.Split(id, "_")[0]
	if split == "avatar" {
		collection = avatarPopularityCollection
	}
	if split == "world" {
		collection = worldPopularityCollection
	}
	if collection == nil {
		return nil
	}
	filter := bson.M{
		"Id": id,
	}
	var result api.PopularityObject
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Create a new one
			popularity := api.CreatePopularity(id)
			_, insertErr := collection.InsertOne(context.TODO(), popularity)
			if insertErr != nil {
				panic(insertErr)
			}
			return &popularity
		} else {
			panic(err)
		}
	}
	return &result
}

func UpdatePopularity(popularity *api.PopularityObject) {
	var collection *mongo.Collection
	split := strings.Split(popularity.Id, "_")[0]
	if split == "avatar" {
		collection = avatarPopularityCollection
	}
	if split == "world" {
		collection = worldPopularityCollection
	}
	if collection == nil {
		return
	}
	filter := bson.D{{"Id", popularity.Id}}
	update := bson.D{{"$set", popularity}}
	_, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		panic(err)
	}
}
