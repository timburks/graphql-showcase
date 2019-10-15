// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	showcase "github.com/googleapis/gapic-showcase/client"
	genprotopb "github.com/googleapis/gapic-showcase/server/genproto"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc"
)

const ADDRESS = "localhost:7469"

func newConnectionOptions() ([]option.ClientOption, error) {
	var opts []option.ClientOption
	conn, err := grpc.Dial(ADDRESS, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	opts = append(opts, option.WithGRPCConn(conn))
	return opts, nil
}

func newEchoClient(ctx context.Context) (*showcase.EchoClient, error) {
	opts, err := newConnectionOptions()
	if err != nil {
		return nil, err
	}
	return showcase.NewEchoClient(ctx, opts...)
}

func newIdentityClient(ctx context.Context) (*showcase.IdentityClient, error) {
	opts, err := newConnectionOptions()
	if err != nil {
		return nil, err
	}
	return showcase.NewIdentityClient(ctx, opts...)
}

var echoType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Echo",
		Fields: graphql.Fields{
			"content": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var userType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"display_name": &graphql.Field{
				Type: graphql.String,
			},
			"email": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

func representationForUser(user *genprotopb.User) map[string]interface{} {
	return map[string]interface{}{
		"id":           user.Name,
		"display_name": user.DisplayName,
		"email":        user.Email,
	}
}

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"echo": &graphql.Field{
				Type: echoType,
				Args: graphql.FieldConfigArgument{
					"request": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					log.Printf("echo(%+v)", p.Args)
					request, isFound := p.Args["request"].(string)
					if !isFound {
						return nil, errors.New("request not found")
					}
					ctx := context.TODO()
					c, err := newEchoClient(ctx)
					if err != nil {
						return nil, err
					}
					req := &genprotopb.EchoRequest{
						Response: &genprotopb.EchoRequest_Content{Content: request},
					}
					log.Printf("request %+v", req)
					resp, err := c.Echo(ctx, req)
					if err != nil {
						return nil, err
					}
					log.Printf("response %+v", resp)
					return resp, nil
				},
			},
			"user": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					log.Printf("user(%+v)", p.Args)
					name, isFound := p.Args["id"].(string)
					if !isFound {
						return nil, errors.New("Arg not found")
					}
					ctx := context.TODO()
					c, err := newIdentityClient(ctx)
					if err != nil {
						return nil, err
					}
					req := &genprotopb.GetUserRequest{
						Name: name,
					}
					log.Printf("request %+v", req)
					user, err := c.GetUser(ctx, req)
					if err != nil {
						return nil, err
					}
					log.Printf("response %+v", user)
					if err != nil {
						return nil, err
					}
					return representationForUser(user), nil
				},
			},
			"users": &graphql.Field{
				Type: graphql.NewList(userType),
				Args: nil,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					log.Printf("users(%+v)", p.Args)
					ctx := context.TODO()
					c, err := newIdentityClient(ctx)
					if err != nil {
						return nil, err
					}
					req := &genprotopb.ListUsersRequest{}
					log.Printf("request %+v", req)
					it := c.ListUsers(ctx, req)
					users := []map[string]interface{}{}
					for {
						user, err := it.Next()
						if err == iterator.Done {
							break
						}
						if err != nil {
							return nil, err
						}
						users = append(users, representationForUser(user))
					}
					log.Printf("response %+v", users)
					return users, nil
				},
			},
		},
	})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"createUser": &graphql.Field{
			Type: userType,
			Args: graphql.FieldConfigArgument{
				"display_name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"email": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				log.Printf("createUser(%+v)", p.Args)
				display_name, isFound := p.Args["display_name"].(string)
				if !isFound {
					return nil, errors.New("Arg not found")
				}
				email, isFound := p.Args["email"].(string)
				if !isFound {
					return nil, errors.New("Arg not found")
				}
				ctx := context.TODO()
				c, err := newIdentityClient(ctx)
				if err != nil {
					return nil, err
				}
				req := &genprotopb.CreateUserRequest{
					User: &genprotopb.User{DisplayName: display_name, Email: email},
				}
				log.Printf("request %+v", req)
				user, err := c.CreateUser(ctx, req)
				if err != nil {
					return nil, err
				}
				log.Printf("response %+v", user)
				return representationForUser(user), err
			},
		},
		"deleteUser": &graphql.Field{
			Type: userType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				log.Printf("deleteUser(%+v)", p.Args)
				name, isFound := p.Args["id"].(string)
				if !isFound {
					return nil, errors.New("Arg not found")
				}
				ctx := context.TODO()
				c, err := newIdentityClient(ctx)
				if err != nil {
					return nil, err
				}
				req := &genprotopb.GetUserRequest{Name: name}
				log.Printf("get request %+v", req)
				user, err := c.GetUser(ctx, req)
				if err != nil {
					return nil, err
				}
				log.Printf("get response %+v", user)
				req2 := &genprotopb.DeleteUserRequest{Name: name}
				log.Printf("delete request %+v", req)
				err = c.DeleteUser(ctx, req2)
				if err != nil {
					return nil, err
				}
				return representationForUser(user), nil
			},
		},
		"updateUser": &graphql.Field{
			Type: userType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"display_name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"email": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				log.Printf("updateUser(%+v)", p.Args)
				name, isFound := p.Args["id"].(string)
				if !isFound {
					return nil, errors.New("Arg not found")
				}
				mask := &field_mask.FieldMask{Paths: make([]string, 0)}
				display_name, isFound := p.Args["display_name"].(string)
				if isFound {
					mask.Paths = append(mask.Paths, "display_name")
				}
				email, isFound := p.Args["email"].(string)
				if isFound {
					mask.Paths = append(mask.Paths, "email")
				}
				user := &genprotopb.User{
					Name:        name,
					DisplayName: display_name,
					Email:       email,
				}
				ctx := context.TODO()
				c, err := newIdentityClient(ctx)
				if err != nil {
					return nil, err
				}
				req := &genprotopb.UpdateUserRequest{
					User:       user,
					UpdateMask: mask,
				}
				req.UpdateMask = nil // Oops. Removing field masks because they are currently not supported by the Identity service.
				log.Printf("request %+v", req)
				updatedUser, err := c.UpdateUser(ctx, req)
				if err != nil {
					return nil, err
				}
				log.Printf("response %+v", req)
				return representationForUser(updatedUser), err
			},
		},
	},
})

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	},
)

func main() {
	// graphql handler
	h := handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	})
	http.Handle("/graphql", h)

	// static file server for Graphiql in-browser editor
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	fmt.Println("Running server on port 8080")
	http.ListenAndServe(":8080", nil)
}
