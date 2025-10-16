package server

var ctScoket = connectionType{}

func redisRegisterConnectionTypeSocket() error {

	return connTypeRegister(&ctScoket)
}
