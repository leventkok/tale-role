package mongostore

import (
	"context"
	"strings"
	"time"

	"github.com/leventkok/tale-role/apps/api/internal/application/game"
	"github.com/leventkok/tale-role/apps/api/internal/application/world"
	"github.com/leventkok/tale-role/apps/api/internal/domain/iam"
	"github.com/leventkok/tale-role/apps/api/internal/domain/license"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct {
	users     *mongo.Collection
	otps      *mongo.Collection
	licenses  *mongo.Collection
	rooms     *mongo.Collection
	universes *mongo.Collection
	packs     *mongo.Collection
}

func Connect(ctx context.Context, uri, dbName string) (*Store, error) {
	if dbName == "" {
		dbName = "talerole"
	}
	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := cli.Ping(ctx, nil); err != nil {
		_ = cli.Disconnect(ctx)
		return nil, err
	}
	db := cli.Database(dbName)
	s := &Store{
		users:     db.Collection("users"),
		otps:      db.Collection("otps"),
		licenses:  db.Collection("licenses"),
		rooms:     db.Collection("rooms"),
		universes: db.Collection("universes"),
		packs:     db.Collection("prompt_packs"),
	}
	_, _ = s.users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return s, nil
}

func (s *Store) PutUser(u *iam.User) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = s.users.ReplaceOne(ctx, bson.M{"_id": u.ID}, userDoc{
		ID: u.ID, Email: strings.ToLower(u.Email), PasswordHash: u.PasswordHash,
		Verified: u.Verified, TOTPSecret: u.TOTPSecret, TOTPEnabled: u.TOTPEnabled,
		LanternXP: u.LanternXP, LanternLevel: u.LanternLevel, PortraitID: u.PortraitID, CreatedAt: u.CreatedAt,
	}, options.Replace().SetUpsert(true))
}

func (s *Store) GetUser(email string) (*iam.User, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var d userDoc
	err := s.users.FindOne(ctx, bson.M{"email": strings.ToLower(email)}).Decode(&d)
	if err != nil {
		return nil, false
	}
	return d.user(), true
}

func (s *Store) GetUserByID(id string) (*iam.User, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var d userDoc
	err := s.users.FindOne(ctx, bson.M{"_id": id}).Decode(&d)
	if err != nil {
		return nil, false
	}
	return d.user(), true
}

func (s *Store) PutOTP(o *iam.OTP) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	email := strings.ToLower(o.Email)
	_, _ = s.otps.ReplaceOne(ctx, bson.M{"_id": email}, otpDoc{
		Email: email, Hash: o.Hash, ExpiresAt: o.ExpiresAt, Attempts: o.Attempts,
	}, options.Replace().SetUpsert(true))
}

func (s *Store) GetOTP(email string) (*iam.OTP, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var d otpDoc
	err := s.otps.FindOne(ctx, bson.M{"_id": strings.ToLower(email)}).Decode(&d)
	if err != nil || time.Now().After(d.ExpiresAt) {
		return nil, false
	}
	return &iam.OTP{Email: d.Email, Hash: d.Hash, ExpiresAt: d.ExpiresAt, Attempts: d.Attempts}, true
}

func (s *Store) DeleteOTP(email string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = s.otps.DeleteOne(ctx, bson.M{"_id": strings.ToLower(email)})
}

func (s *Store) PutLicense(l *license.ProductLicense) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = s.licenses.ReplaceOne(ctx, bson.M{"_id": l.ID}, licDoc{
		ID: l.ID, UserID: l.UserID, DeviceID: l.DeviceID, Platform: l.Platform, CreatedAt: l.CreatedAt,
	}, options.Replace().SetUpsert(true))
}

func (s *Store) DeleteLicense(id string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = s.licenses.DeleteOne(ctx, bson.M{"_id": id})
}

func (s *Store) LicensesForUser(userID string) []*license.ProductLicense {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := s.licenses.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return []*license.ProductLicense{}
	}
	defer cur.Close(ctx)
	out := []*license.ProductLicense{}
	for cur.Next(ctx) {
		var d licDoc
		if cur.Decode(&d) != nil {
			continue
		}
		out = append(out, &license.ProductLicense{
			ID: d.ID, UserID: d.UserID, DeviceID: d.DeviceID, Platform: d.Platform, CreatedAt: d.CreatedAt,
		})
	}
	return out
}

func (s *Store) DeleteUserByID(id string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	u, ok := s.GetUserByID(id)
	if ok {
		_, _ = s.otps.DeleteOne(ctx, bson.M{"_id": strings.ToLower(u.Email)})
	}
	_, _ = s.users.DeleteOne(ctx, bson.M{"_id": id})
	_, _ = s.licenses.DeleteMany(ctx, bson.M{"user_id": id})
}

func (s *Store) UpsertRoom(r *game.Room) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.rooms.ReplaceOne(ctx, bson.M{"_id": r.ID}, encodeRoom(r), options.Replace().SetUpsert(true))
	return err
}

func (s *Store) DeleteRoom(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.rooms.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (s *Store) LoadRooms(ctx context.Context) ([]*game.Room, error) {
	cur, err := s.rooms.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []*game.Room{}
	for cur.Next(ctx) {
		var d roomDoc
		if cur.Decode(&d) != nil {
			continue
		}
		out = append(out, decodeRoom(d))
	}
	return out, nil
}

func (s *Store) UpsertUniverse(u *world.Universe) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.universes.ReplaceOne(ctx, bson.M{"_id": u.ID}, u, options.Replace().SetUpsert(true))
	return err
}

func (s *Store) DeleteUniverse(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.universes.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (s *Store) LoadUniverses(ctx context.Context) ([]*world.Universe, error) {
	cur, err := s.universes.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []*world.Universe{}
	for cur.Next(ctx) {
		var u world.Universe
		if cur.Decode(&u) != nil {
			continue
		}
		cp := u
		out = append(out, &cp)
	}
	return out, nil
}

type PromptPack struct {
	ID string `bson:"_id"`
	EN string `bson:"en"`
	TR string `bson:"tr"`
}

func (s *Store) SavePromptPack(id, en, tr string) error {
	if s == nil || s.packs == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.packs.ReplaceOne(ctx, bson.M{"_id": id}, PromptPack{ID: id, EN: en, TR: tr}, options.Replace().SetUpsert(true))
	return err
}

func (s *Store) LoadPromptPacks() []PromptPack {
	if s == nil || s.packs == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := s.packs.Find(ctx, bson.M{})
	if err != nil {
		return nil
	}
	defer cur.Close(ctx)
	out := []PromptPack{}
	for cur.Next(ctx) {
		var d PromptPack
		if cur.Decode(&d) != nil {
			continue
		}
		out = append(out, d)
	}
	return out
}
