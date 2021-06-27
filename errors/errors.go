package errors

type customError string

func (err customError) Error() string {
	return string(err)
}

const BoardBusyError customError = "Board is currently busy"
const BoardReadyError customError = "Board is ready to accept commands"
