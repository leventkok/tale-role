package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/leventkok/tale-role/apps/api/internal/application/game"
	"github.com/leventkok/tale-role/apps/api/internal/application/world"
	"github.com/leventkok/tale-role/apps/api/internal/domain/iam"
	"github.com/leventkok/tale-role/apps/api/internal/shared/httperr"
	gateway "github.com/leventkok/tale-role/services/llm-gateway"
)

func gqlUser(p graphql.ResolveParams) *iam.User {
	u, _ := p.Context.Value(userKey).(*iam.User)
	return u
}

func (s *Server) graphQLSchema() (graphql.Schema, error) {
	memberType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Member",
		Fields: graphql.Fields{
			"userId": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"role":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	})
	statsType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Stats",
		Fields: graphql.Fields{
			"str": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"dex": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"con": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"int": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"wis": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"cha": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		},
	})
	characterType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Character",
		Fields: graphql.Fields{
			"userId":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"name":          &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"species":       &graphql.Field{Type: graphql.String},
			"path":          &graphql.Field{Type: graphql.String},
			"backstory":     &graphql.Field{Type: graphql.String},
			"skills":        &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
			"hp":            &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"xp":            &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"level":         &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"initiative":    &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"hasInitiative": &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"stats":         &graphql.Field{Type: statsType},
		},
	})
	sceneType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Scene",
		Fields: graphql.Fields{
			"themeId":      &graphql.Field{Type: graphql.String},
			"visualPrompt": &graphql.Field{Type: graphql.String},
			"imageSvg":     &graphql.Field{Type: graphql.String},
			"inference":    &graphql.Field{Type: graphql.String},
		},
	})
	turnType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Turn",
		Fields: graphql.Fields{
			"actorId": &graphql.Field{Type: graphql.String},
			"kind":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"notes":   &graphql.Field{Type: graphql.String},
			"rolls":   &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.Int)))},
			"total":   &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"success": &graphql.Field{Type: graphql.Boolean},
			"prose":   &graphql.Field{Type: graphql.String},
		},
	})
	roomType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Room",
		Fields: graphql.Fields{
			"id":                &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"name":              &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"hostId":            &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"diceSystem":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"joinMode":          &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"started":           &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"completed":         &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"universeId":        &graphql.Field{Type: graphql.String},
			"themeId":           &graphql.Field{Type: graphql.String},
			"promptPackVersion": &graphql.Field{Type: graphql.String},
			"turnOrder":         &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
			"currentActorId":    &graphql.Field{Type: graphql.String},
			"presence":          &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(memberType)))},
			"characters":        &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(characterType)))},
			"turns":             &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(turnType)))},
			"scene":             &graphql.Field{Type: sceneType},
		},
	})
	universeSummary := graphql.NewObject(graphql.ObjectConfig{
		Name: "UniverseSummary",
		Fields: graphql.Fields{
			"id":                &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"nameEn":            &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"themeId":           &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"diceSystem":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"promptPackVersion": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	})
	universeType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Universe",
		Fields: graphql.Fields{
			"id":             &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"nameEn":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"themeId":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"diceSystem":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"compiledPrompt": &graphql.Field{Type: graphql.String},
		},
	})
	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id":           &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"email":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"verified":     &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"totpEnabled":  &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"lanternXp":    &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"lanternLevel": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"portraitId":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	})
	healthType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Health",
		Fields: graphql.Fields{
			"status":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"persistence": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"llm":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"images":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"mail":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	})
	createdType := graphql.NewObject(graphql.ObjectConfig{
		Name: "RoomCreated",
		Fields: graphql.Fields{
			"id":         &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"diceSystem": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"universeId": &graphql.Field{Type: graphql.String},
		},
	})
	lobbyType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Lobby",
		Fields: graphql.Fields{
			"id":       &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"name":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"joinMode": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"started":  &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"seats":    &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		},
	})
	licenseType := graphql.NewObject(graphql.ObjectConfig{
		Name: "License",
		Fields: graphql.Fields{
			"id":        &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"deviceId":  &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"platform":  &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"createdAt": &graphql.Field{Type: graphql.String},
		},
	})
	statsInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "StatsInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"str": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
			"dex": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
			"con": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
			"int": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
			"wis": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
			"cha": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
		},
	})
	npcInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "NPCInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"nameEn":    &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"nameTr":    &graphql.InputObjectFieldConfig{Type: graphql.String},
			"alignment": &graphql.InputObjectFieldConfig{Type: graphql.String},
			"voice":     &graphql.InputObjectFieldConfig{Type: graphql.String},
		},
	})

	query := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"health": &graphql.Field{
				Type: graphql.NewNonNull(healthType),
				Resolve: func(graphql.ResolveParams) (any, error) {
					rt := s.llm.Runtime()
					return map[string]any{
						"status": "ready", "persistence": s.cfg.Persistence(),
						"llm": rt.Inference, "images": "stub", "mail": s.cfg.Mail(),
					}, nil
				},
			},
			"me": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, nil
					}
					lvl := u.LanternLevel
					if lvl < 1 {
						lvl = 1
					}
					return map[string]any{
						"id": u.ID, "email": u.Email, "verified": u.Verified, "totpEnabled": u.TOTPEnabled,
						"lanternXp": u.LanternXP, "lanternLevel": lvl, "portraitId": iam.NormalizePortrait(u.PortraitID),
					}, nil
				},
			},
			"room": &graphql.Field{
				Type: roomType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					id, _ := p.Args["id"].(string)
					pub, err := s.table.View(id, u.ID)
					if err != nil {
						return nil, err
					}
					return roomMap(pub), nil
				},
			},
			"universes": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(universeSummary))),
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					rows := s.worlds.List(u.ID)
					out := make([]map[string]any, 0, len(rows))
					for _, row := range rows {
						out = append(out, map[string]any{
							"id": row.ID, "nameEn": row.NameEN, "themeId": row.ThemeID,
							"diceSystem": row.DiceSystem, "promptPackVersion": row.PromptPackVersion,
						})
					}
					return out, nil
				},
			},
			"universe": &graphql.Field{
				Type: universeType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					id, _ := p.Args["id"].(string)
					doc, err := s.worlds.Get(id, u.ID)
					if err != nil {
						return nil, err
					}
					return map[string]any{
						"id": doc.ID, "nameEn": doc.NameEN, "themeId": doc.ThemeID,
						"diceSystem": doc.DiceSystem, "compiledPrompt": doc.CompiledPrompt,
					}, nil
				},
			},
			"lobbies": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(lobbyType))),
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					rows := s.table.Lobbies()
					out := make([]map[string]any, 0, len(rows))
					for _, row := range rows {
						out = append(out, map[string]any{
							"id": row.ID, "name": row.Name, "joinMode": row.JoinMode,
							"started": row.Started, "seats": row.Seats,
						})
					}
					return out, nil
				},
			},
			"licenses": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(licenseType))),
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					list := s.svc.Licenses(u.ID)
					out := make([]map[string]any, 0, len(list))
					for _, l := range list {
						out = append(out, map[string]any{
							"id": l.ID, "deviceId": l.DeviceID, "platform": l.Platform,
							"createdAt": l.CreatedAt.UTC().Format(time.RFC3339),
						})
					}
					return out, nil
				},
			},
		},
	})

	mutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createRoom": &graphql.Field{
				Type: createdType,
				Args: graphql.FieldConfigArgument{
					"name":       &graphql.ArgumentConfig{Type: graphql.String},
					"joinMode":   &graphql.ArgumentConfig{Type: graphql.String},
					"password":   &graphql.ArgumentConfig{Type: graphql.String},
					"diceSystem": &graphql.ArgumentConfig{Type: graphql.String},
					"universeId": &graphql.ArgumentConfig{Type: graphql.ID},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					name, _ := p.Args["name"].(string)
					join, _ := p.Args["joinMode"].(string)
					pass, _ := p.Args["password"].(string)
					dice, _ := p.Args["diceSystem"].(string)
					uniID, _ := p.Args["universeId"].(string)
					if uniID != "" {
						uni, err := s.worlds.GetForHost(uniID, u.ID)
						if err != nil {
							return nil, err
						}
						dice = uni.DiceSystem
						if strings.TrimSpace(name) == "" {
							name = uni.NameEN
						}
						room, err := s.table.Create(u.ID, name, join, pass, dice)
						if err != nil {
							return nil, err
						}
						_ = s.table.BindUniverse(room.ID, uni.ID, uni.ThemeID, uni.PromptPackVersion)
						s.seatSavedHero(room.ID, u.ID)
						return map[string]any{"id": room.ID, "diceSystem": room.DiceSystem, "universeId": uni.ID}, nil
					}
					room, err := s.table.Create(u.ID, name, join, pass, dice)
					if err != nil {
						return nil, err
					}
					return map[string]any{"id": room.ID, "diceSystem": room.DiceSystem}, nil
				},
			},
			"joinRoom": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Args: graphql.FieldConfigArgument{
					"roomId":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"password": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					roomID, _ := p.Args["roomId"].(string)
					pass, _ := p.Args["password"].(string)
					role := "player"
					if u.Email != "" && s.adminEmail != "" && u.Email == s.adminEmail {
						role = "system_admin"
					}
					if err := s.table.Join(roomID, u.ID, pass, role); err != nil {
						return nil, err
					}
					s.seatSavedHero(roomID, u.ID)
					return true, nil
				},
			},
			"startRoom": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Args: graphql.FieldConfigArgument{
					"roomId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"locale": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					roomID, _ := p.Args["roomId"].(string)
					if err := s.table.Start(roomID, u.ID); err != nil {
						return nil, err
					}
					locale := gqlString(p.Args["locale"])
					if locale == "" {
						locale = "en"
					}
					if pub, err := s.table.View(roomID, u.ID); err == nil && len(pub.Turns) > 0 {
						last := pub.Turns[len(pub.Turns)-1]
						_ = s.narrateTurn(roomID, u.ID, locale, last.Notes, last)
					}
					return true, nil
				},
			},
			"completeRoom": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Args: graphql.FieldConfigArgument{
					"roomId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					roomID, _ := p.Args["roomId"].(string)
					ids, err := s.table.Complete(roomID, u.ID)
					if err != nil {
						return nil, err
					}
					for _, id := range ids {
						s.svc.GrantLantern(id, game.TaleCompleteXP)
					}
					return true, nil
				},
			},
			"rollInitiative": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
				Args: graphql.FieldConfigArgument{
					"roomId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					roomID, _ := p.Args["roomId"].(string)
					n, err := s.table.RollInitiative(roomID, u.ID)
					if err != nil {
						return nil, err
					}
					return n, nil
				},
			},
			"setCharacter": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Args: graphql.FieldConfigArgument{
					"roomId":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"name":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"stats":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(statsInput)},
					"species":   &graphql.ArgumentConfig{Type: graphql.String},
					"path":      &graphql.ArgumentConfig{Type: graphql.String},
					"backstory": &graphql.ArgumentConfig{Type: graphql.String},
					"skills":    &graphql.ArgumentConfig{Type: graphql.NewList(graphql.NewNonNull(graphql.String))},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					roomID, _ := p.Args["roomId"].(string)
					if err := s.table.SetSheet(roomID, u.ID, game.Sheet{
						Name:      gqlString(p.Args["name"]),
						Species:   gqlString(p.Args["species"]),
						Path:      gqlString(p.Args["path"]),
						Backstory: gqlString(p.Args["backstory"]),
						Stats:     statsFromGQL(gqlMap(p.Args["stats"])),
						Skills:    stringList(p.Args["skills"]),
					}); err != nil {
						return nil, err
					}
					s.rememberHero(roomID, u.ID)
					return true, nil
				},
			},
			"actTurn": &graphql.Field{
				Type: turnType,
				Args: graphql.FieldConfigArgument{
					"roomId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"kind":   &graphql.ArgumentConfig{Type: graphql.String},
					"skill":  &graphql.ArgumentConfig{Type: graphql.String},
					"notes":  &graphql.ArgumentConfig{Type: graphql.String},
					"dc":     &graphql.ArgumentConfig{Type: graphql.Int},
					"locale": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					roomID, _ := p.Args["roomId"].(string)
					kind, _ := p.Args["kind"].(string)
					skill, _ := p.Args["skill"].(string)
					notes, _ := p.Args["notes"].(string)
					locale, _ := p.Args["locale"].(string)
					dc := gqlInt(p.Args["dc"])
					_ = s.llm.ProposeIntent(gateway.IntentRequest{
						Locale: locale, RoomID: roomID, Kind: kind, Skill: skill, Notes: notes,
					})
					turn, err := s.table.Act(roomID, u.ID, kind, skill, notes, dc)
					if err != nil {
						return nil, err
					}
					turn = s.narrateTurn(roomID, u.ID, locale, notes, turn)
					s.rememberHero(roomID, u.ID)
					prose := ""
					if turn.Narrative != nil {
						prose = turn.Narrative.Prose
					}
					return map[string]any{
						"actorId": turn.ActorID, "kind": turn.Kind, "notes": turn.Notes,
						"rolls": turnRolls(turn.Rolls), "total": turn.Total, "success": turn.Success, "prose": prose,
					}, nil
				},
			},
			"createUniverse": &graphql.Field{
				Type: universeType,
				Args: graphql.FieldConfigArgument{
					"nameEn":        &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"nameTr":        &graphql.ArgumentConfig{Type: graphql.String},
					"themeId":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"diceSystem":    &graphql.ArgumentConfig{Type: graphql.String},
					"contentRating": &graphql.ArgumentConfig{Type: graphql.String},
					"era":           &graphql.ArgumentConfig{Type: graphql.String},
					"tone":          &graphql.ArgumentConfig{Type: graphql.String},
					"description":   &graphql.ArgumentConfig{Type: graphql.String},
					"opening":       &graphql.ArgumentConfig{Type: graphql.String},
					"taboos":        &graphql.ArgumentConfig{Type: graphql.String},
					"npcs":          &graphql.ArgumentConfig{Type: graphql.NewList(npcInput)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					nameEn, _ := p.Args["nameEn"].(string)
					themeID, _ := p.Args["themeId"].(string)
					doc, err := s.worlds.Create(u.ID, world.Draft{
						NameEN:        nameEn,
						NameTR:        gqlString(p.Args["nameTr"]),
						ThemeID:       themeID,
						DiceSystem:    gqlString(p.Args["diceSystem"]),
						ContentRating: gqlString(p.Args["contentRating"]),
						Era:           gqlString(p.Args["era"]),
						Tone:          gqlString(p.Args["tone"]),
						Description:   gqlString(p.Args["description"]),
						Opening:       gqlString(p.Args["opening"]),
						Taboos:        gqlString(p.Args["taboos"]),
						NPCs:          npcsFromGQL(p.Args["npcs"]),
					})
					if err != nil {
						return nil, err
					}
					s.svc.GrantLantern(u.ID, 25)
					return map[string]any{
						"id": doc.ID, "nameEn": doc.NameEN, "themeId": doc.ThemeID,
						"diceSystem": doc.DiceSystem, "compiledPrompt": doc.CompiledPrompt,
					}, nil
				},
			},
			"registerLicense": &graphql.Field{
				Type: licenseType,
				Args: graphql.FieldConfigArgument{
					"deviceId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"platform": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					deviceID, _ := p.Args["deviceId"].(string)
					platform, _ := p.Args["platform"].(string)
					lic, err := s.svc.RegisterLicense(u.ID, deviceID, platform)
					if err != nil {
						return nil, err
					}
					return map[string]any{
						"id": lic.ID, "deviceId": lic.DeviceID, "platform": lic.Platform,
						"createdAt": lic.CreatedAt.UTC().Format(time.RFC3339),
					}, nil
				},
			},
			"setPortrait": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					id, _ := p.Args["id"].(string)
					if err := s.svc.SetPortrait(u.ID, id); err != nil {
						return nil, err
					}
					return true, nil
				},
			},
			"eraseMe": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := gqlUser(p)
					if u == nil {
						return nil, fmt.Errorf("unauthorized")
					}
					s.table.ForgetUser(u.ID)
					s.worlds.ForgetOwner(u.ID)
					s.worlds.ForgetPlayer(u.ID)
					if err := s.svc.Erase(u.ID); err != nil {
						return nil, err
					}
					return true, nil
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{Query: query, Mutation: mutation})
}

func roomMap(pub *game.PublicRoom) map[string]any {
	presence := make([]map[string]any, 0, len(pub.Presence))
	for _, m := range pub.Presence {
		presence = append(presence, map[string]any{"userId": m.UserID, "role": m.Role})
	}
	chars := make([]map[string]any, 0, len(pub.Characters))
	for _, ch := range pub.Characters {
		skills := ch.Skills
		if skills == nil {
			skills = []string{}
		}
		chars = append(chars, map[string]any{
			"userId": ch.UserID, "name": ch.Name, "species": ch.Species, "path": ch.Path, "backstory": ch.Backstory,
			"skills": skills, "hp": ch.HP,
			"xp": ch.XP, "level": ch.Level, "initiative": ch.Initiative, "hasInitiative": ch.HasInitiative,
			"stats": map[string]any{
				"str": ch.Stats.STR, "dex": ch.Stats.DEX, "con": ch.Stats.CON,
				"int": ch.Stats.INT, "wis": ch.Stats.WIS, "cha": ch.Stats.CHA,
			},
		})
	}
	turns := make([]map[string]any, 0, len(pub.Turns))
	for _, t := range pub.Turns {
		prose := ""
		if t.Narrative != nil {
			prose = t.Narrative.Prose
		}
		turns = append(turns, map[string]any{
			"actorId": t.ActorID, "kind": t.Kind, "notes": t.Notes,
			"rolls": turnRolls(t.Rolls), "total": t.Total, "success": t.Success, "prose": prose,
		})
	}
	order := pub.TurnOrder
	if order == nil {
		order = []string{}
	}
	out := map[string]any{
		"id": pub.ID, "name": pub.Name, "hostId": pub.HostID, "diceSystem": pub.DiceSystem,
		"joinMode": pub.JoinMode, "started": pub.Started, "completed": pub.Completed, "universeId": pub.UniverseID,
		"themeId": pub.ThemeID, "promptPackVersion": pub.PromptPackVersion,
		"turnOrder": order, "currentActorId": pub.CurrentActorID,
		"presence": presence, "characters": chars, "turns": turns,
	}
	if pub.Scene != nil {
		out["scene"] = map[string]any{
			"themeId": pub.Scene.ThemeID, "visualPrompt": pub.Scene.VisualPrompt,
			"imageSvg": pub.Scene.ImageSVG, "inference": pub.Scene.Inference,
		}
	}
	return out
}

func turnRolls(rolls []int) []int {
	if rolls == nil {
		return []int{}
	}
	return rolls
}

func gqlString(v any) string {
	s, _ := v.(string)
	return s
}

func gqlMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func stringList(v any) []string {
	rows, _ := v.([]any)
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		s, _ := row.(string)
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

func gqlInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int32:
		return int(n)
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

func statsFromGQL(raw map[string]any) game.Stats {
	if raw == nil {
		return game.Stats{}
	}
	return game.Stats{
		STR: gqlInt(raw["str"]), DEX: gqlInt(raw["dex"]), CON: gqlInt(raw["con"]),
		INT: gqlInt(raw["int"]), WIS: gqlInt(raw["wis"]), CHA: gqlInt(raw["cha"]),
	}
}

func npcsFromGQL(v any) []world.NPC {
	rows, _ := v.([]any)
	out := make([]world.NPC, 0, len(rows))
	for _, row := range rows {
		m, ok := row.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, world.NPC{
			NameEN: gqlString(m["nameEn"]), NameTR: gqlString(m["nameTr"]),
			Alignment: gqlString(m["alignment"]), Voice: gqlString(m["voice"]),
		})
	}
	return out
}

func (s *Server) optionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			next.ServeHTTP(w, r)
			return
		}
		if !strings.HasPrefix(header, "Bearer ") {
			httperr.Write(w, s.log, http.StatusUnauthorized, "unauthorized", nil)
			return
		}
		u, err := s.svc.UserFromToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			httperr.Write(w, s.log, http.StatusUnauthorized, "unauthorized", nil)
			return
		}
		next.ServeHTTP(w, r.WithContext(withUser(r.Context(), u)))
	})
}

func (s *Server) graphQL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query         string         `json:"query"`
		Variables     map[string]any `json:"variables"`
		OperationName string         `json:"operationName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	result := graphql.Do(graphql.Params{
		Schema:         s.gql,
		RequestString:  req.Query,
		VariableValues: req.Variables,
		OperationName:  req.OperationName,
		Context:        r.Context(),
	})
	httperr.JSON(w, http.StatusOK, result)
}
