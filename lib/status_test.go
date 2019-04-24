package lib

import "fmt"

var values = []Status{
	UNDEFINED,
	STARTED,
	SUCCEEDED,
	FAILED,
	Status(50),
}

func ExampleStatus_String() {
	for _, a := range values {
		fmt.Println(a)
	}

	// Output:
	// undefined
	// started
	// succeeded
	// failed
	// unknown-status-50
}

func ExampleStatus_Includes() {
	for _, a := range values {
		for _, b := range values {
			if a.Includes(b) {
				fmt.Println(a, "includes", b)
			}
		}
	}

	// Output:
	// started includes started
	// succeeded includes started
	// succeeded includes succeeded
	// failed includes started
	// failed includes failed
}

func ExampleStatus_MarshalText() {
	for _, a := range values {
		if bs, err := a.MarshalText(); err == nil {
			fmt.Println(string(bs))
		} else {
			fmt.Println(err.Error())
		}
	}

	// Output:
	// undefined
	// started
	// succeeded
	// failed
	// unknown status #50
}

func ExampleStatus_UnmarshalText() {
	texts := []string{
		"undefined",
		" UndefinED ",
		"started",
		"STARTED",
		"SUCCEEDed",
		"faileD ",
		"rouge",
	}

	for _, a := range texts {
		var status Status
		if err := status.UnmarshalText([]byte(a)); err == nil {
			fmt.Println(status)
		} else {
			fmt.Println(err.Error())
		}
	}

	// Output:
	// undefined
	// undefined
	// started
	// started
	// succeeded
	// failed
	// unknown status: rouge
}
