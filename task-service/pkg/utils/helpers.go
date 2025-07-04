package utils

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type Envelope map[string]interface{}


func WriteJSON(w http.ResponseWriter, r *http.Request, status int, data Envelope, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func ReadJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error{
	err := json.NewDecoder(r.Body).Decode(&dst)
	if err != nil {
		var sytaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &sytaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", sytaxError.Offset)
		

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			return fmt.Errorf("body contains an invalid value for the %q field (at character %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// uncomment this to set maxBytes for the request body to prevent DoS attacks
		// case err.Error() == "http: request body too large":
		// 	return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
			

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
	}

	}
	return nil
}

func GetURLparams(r *http.Request, name string) string {
	params := httprouter.ParamsFromContext(r.Context())
	param := params.ByName(name)
	return param
}

func ConvertIDtoNum(w http.ResponseWriter, id string) (int, error) {
	id = strings.ReplaceAll(id, " ", "")
	if len(id) > 0 { // Ensure the string is not empty
		if id[0] == ':' {
			// id = id[1:]
			id = strings.ReplaceAll(id, ":", "")
		}
	}
	user_id, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}
	return user_id, err
}

func IntToUUID(i int64) uuid.UUID {
	var u uuid.UUID
	// Put the int64 into the first 8 bytes of the UUID
	binary.BigEndian.PutUint64(u[:8], uint64(i))
	// The rest of the bytes can be zero, or filled with random values
	return u
}



func CheckContains(list []string, target string) bool {
	for _, v := range list {
			if v == target {
					return true
			}
	}
	return false
}


func FromStringToTime(target string) (*time.Time, error) {
	layout := "2006/01/02"

	parsedTime, err := time.Parse(layout, target)
	if err != nil {
			fmt.Println("Error parsing time:", err)
			return nil, err
	}

	return &parsedTime, nil
}