package tggateway_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/arkriny/tggateway"
)

func Example() {
	ctx := context.Background()
	c := tggateway.Client{
		AccessToken: os.Getenv("TG_GATEWAY_TOKEN"),
	}

	sendRes, err := c.SendVerificationMessage(ctx, &tggateway.SendVerificationMessageParams{
		PhoneNumber: os.Getenv("TG_GATEWAY_PHONE"),
		Code:        "1234",
	})
	if err != nil {
		log.Fatal(err)
	}
	requestID := sendRes.RequestID

	checkRes, err := c.CheckVerificationStatus(ctx, &tggateway.CheckVerificationStatusParams{
		RequestID: requestID,
		Code:      "1234",
	})
	if err != nil {
		log.Fatal(err)
	}

	if checkRes.VerificationStatus.Status != tggateway.VerificationStatusValid {
		log.Fatal("Invalid code")
	}
	fmt.Println("Code is valid!")
	// Output: Code is valid!
}
