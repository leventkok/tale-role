package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/leventkok/tale-role/apps/api/internal/application/game"
	"github.com/leventkok/tale-role/apps/api/internal/domain/iam"
	"github.com/leventkok/tale-role/apps/api/internal/shared/httperr"
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
	characterType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Character",
		Fields: graphql.Fields{
			"userId": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"name":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"hp":     &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
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
			"universeId":        &graphql.Field{Type: graphql.String},
			"themeId":           &graphql.Field{Type: graphql.String},
			"promptPackVersion": &graphql.Field{Type: graphql.String},
			"presence":          &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(memberType)))},
			"characters":        &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(characterType)))},
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
			"id":       &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"email":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"verified": &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
		},
	})
	healthType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Health",
		Fields: graphql.Fields{
			"status":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"persistence": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"llm":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"images":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
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

	query := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"health": &graphql.Field{
				Type: graphql.NewNonNull(healthType),
				Resolve: func(graphql.ResolveParams) (any, error) {
					rt := s.llm.Runtime()
					return map[string]any{
						"status": "ready", "persistence": s.cfg.Persistence(),
						"llm": rt.Inference, "images": "stub",
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
					return map[string]any{"id": u.ID, "email": u.Email, "verified": u.Verified}, nil
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
		chars = append(chars, map[string]any{"userId": ch.UserID, "name": ch.Name, "hp": ch.HP})
	}
	return map[string]any{
		"id": pub.ID, "name": pub.Name, "hostId": pub.HostID, "diceSystem": pub.DiceSystem,
		"joinMode": pub.JoinMode, "started": pub.Started, "universeId": pub.UniverseID,
		"themeId": pub.ThemeID, "promptPackVersion": pub.PromptPackVersion,
		"presence": presence, "characters": chars,
	}
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
