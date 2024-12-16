package main

import (
	"fmt"
	"log"

	"github.com/linkedin/goavro/v2"
)

func main() {
	// Avro schema
	schema := `{
        "type": "record",
        "name": "Order",
        "fields": [
            {"name": "orderId", "type": "string"},
            {"name": "userId", "type": "string"},
            {"name": "items", "type": {"type": "array", "items": "string"}}
        ]
    }`

	// paste you bianry here for decode, make sure it align with the schema above. otherwise, either your schame
	// or binary is wrong and that is a good thing to be told!
	// Avro binary data
	avroBinary := []byte{72, 54, 49, 98, 53, 54, 50, 102, 49, 45, 99, 55, 56, 55, 45, 52, 99, 57, 51, 45, 56, 97, 51, 48, 45, 102, 99, 98, 57, 48, 54, 56, 98, 55, 97, 50, 102, 6, 49, 50, 51, 4, 10, 105, 116, 101, 109, 49, 10, 105, 116, 101, 109, 50, 0}

	// Create Avro codec
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		log.Fatal(err)
	}

	// Decode Avro binary data
	native, _, err := codec.NativeFromBinary(avroBinary)
	if err != nil {
		log.Fatal(err)
	}

	// Print decoded data
	fmt.Printf("Decoded data: %+v\n", native)
}
