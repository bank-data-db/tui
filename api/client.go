package api

import (
	"context"
	"log"
	"strings"

	"github.com/bank-data-db/proto/bank_svc_pb"
	"github.com/bank-data-db/proto/errors_pb"
	"github.com/bank-data-db/proto/user_svc_pb"
	"github.com/bank_data_tui/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Client struct {
	jwt string
	bank_svc_pb.BankDataClient
	usr user_svc_pb.UserServiceClient
}

func (c *Client) Login(username, password string) error {
	resp, err := c.usr.CreateToken(context.Background(), user_svc_pb.ReqLogin_builder{
		Username: new(username),
		Password: new(password),
	}.Build())
	if err != nil {
		return err
	}

	c.jwt = resp.GetToken()

	return nil
}

type ValidationErr struct {
	Errors []*errors_pb.ValidationError
}

func (e *ValidationErr) Error() string {
	fields := &strings.Builder{}
	for _, e := range e.Errors {
		for _, f := range e.GetFields() {
			fields.WriteString(f)
		}
	}

	return fields.String()
}

func NewClient(trg string) (*Client, error) {
	c := &Client{}
	conn, err := grpc.NewClient(
		trg,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			if c.jwt != "" {
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", c.jwt)
			}

			err := invoker(ctx, method, req, reply, cc, opts...)
			if err != nil {
				s, ok := status.FromError(err)
				if ok {
					log.Printf("GRPC Error on %v: %v\n", method, err)
					switch s.Code() {
					case codes.Unauthenticated, codes.PermissionDenied:
						if !strings.HasSuffix(method, "CreateToken") {
							utils.GoToScreen(utils.S_LOGIN)
							return nil // special case: is handled by the chan
						}
					case codes.InvalidArgument:
						if len(s.Details()) != 0 {
							details := make([]*errors_pb.ValidationError, 0, len(s.Details()))
							for _, v := range s.Details() {
								d, ok := v.(*errors_pb.ValidationError)
								if ok {
									details = append(details, d)
									log.Println(d)
								}
							}
							if len(details) != 0 {
								return &ValidationErr{details}
							}
						}
					}
				}
			}

			return err
		}),
	)
	if err != nil {
		return nil, err
	}

	c.BankDataClient = bank_svc_pb.NewBankDataClient(conn)
	c.usr = user_svc_pb.NewUserServiceClient(conn)

	return c, nil
}
