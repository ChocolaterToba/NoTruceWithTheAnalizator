package customError

type customError string

func (err customError) Error() string {
	return string(err)
}

const BoardBusyError customError = "Board is currently busy"
const BoardReadyError customError = "Board is ready to accept commands"
const ZoneNotFoundError customError = "Could not find zone"
const SectionNotFoundError customError = "Could not find section in zone"
const ParamNotFoundError customError = "Could not find parameter in section"
const CommandNotFoundError customError = "Could not find requested command"
