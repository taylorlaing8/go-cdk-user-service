package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	logic "cf-user/get-user"
)

func main() {
	if logic.LambdaConfig == nil {
		logic.InitLambda(nil)
	}

	lambda.Start(logic.Handler)
}
