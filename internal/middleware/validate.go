package middleware

func Validate[T any](body T) map[string]any {
	if v, ok := any(body).(interface{ Validate() map[string]any }); ok {
		return v.Validate()
	}
	return nil
}
