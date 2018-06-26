package errors

const (
	JSONParseError          = "Unable to parse request body as JSON"
	JSONError               = "Unable to convert into JSON"
	CreateResourceError     = "Unable to create resource"
	ResourceNotFoundError   = "Unable to find resource"
	DBInsertError           = "Unable to insert into database"
	DBUpdateError           = "Unable to update databse"
	DBGetError              = "Unable to retrieve data from database"
	InvalidPathParamError   = "Received bad bath paramater"
	InvalidCredentialsError = "Invalid username or password"
	InvalidQueryParamError  = "Invalid query paramater"
	InternalServerError     = "Internal server error"
	LoginFailedError        = "Login failed"
	LogoutFailedError       = "Logout failed"
	EmptyCredentialsError   = "Username and password both must be non-empty"
	BadIDError              = "The requested ID does not exist in our system"
)
