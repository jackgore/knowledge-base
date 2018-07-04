package errors

const (
	BadIDError              = "The requested ID does not exist in our system"
	CreateResourceError     = "Unable to create resource"
	DBGetError              = "Unable to retrieve data from database"
	DBInsertError           = "Unable to insert into database"
	DBUpdateError           = "Unable to update databse"
	EmptyCredentialsError   = "Username and password both must be non-empty"
	InternalServerError     = "Internal server error"
	InvalidCredentialsError = "Invalid username or password"
	InvalidPathParamError   = "Received bad bath paramater"
	InvalidQueryParamError  = "Invalid query paramater"
	JSONError               = "Unable to convert into JSON"
	JSONParseError          = "Unable to parse request body as JSON"
	LoginFailedError        = "Login failed"
	LogoutFailedError       = "Logout failed"
	ResourceNotFoundError   = "Unable to find resource"
)
