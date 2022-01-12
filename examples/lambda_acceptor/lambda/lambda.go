package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aidansteele/flowdog/examples/lambda_acceptor"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(handle)
}

func handle(ctx context.Context, input *lambda_acceptor.AcceptorInput) (*lambda_acceptor.AcceptorOutput, error) {
	j, _ := json.Marshal(input)
	fmt.Println(string(j))

	output := &lambda_acceptor.AcceptorOutput{Accept: true}
	j, _ = json.Marshal(output)
	fmt.Println(string(j))
	return output, nil
}
