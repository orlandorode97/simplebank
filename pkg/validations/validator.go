package validations

import (
	"github.com/go-playground/validator/v10"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// validate validates the `input` struct and its validate tag.
func validate(input interface{}) map[string]string {
	validate := validator.New()

	errors := make(map[string]string, 0)
	if err := validate.Struct(input); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			field := err.Field()
			errors[field] = err.Error()
		}
	}

	return errors
}

func convertToBadRequestDetails(errors map[string]string) []*errdetails.BadRequest_FieldViolation {
	details := make([]*errdetails.BadRequest_FieldViolation, 0)
	for field, msg := range errors {
		details = append(details, &errdetails.BadRequest_FieldViolation{
			Field:       field,
			Description: msg,
		})
	}

	return details
}

// BuildErrDetails builds the grpc error status details after validating the validator input.
func BuildErrDetails(validator interface{}, msg string) error {
	errors := validate(validator)
	if len(errors) != 0 {
		st := status.New(codes.InvalidArgument, msg)
		details, _ := st.WithDetails(&errdetails.BadRequest{
			FieldViolations: convertToBadRequestDetails(errors),
		})

		return details.Err()
	}

	return nil
}
