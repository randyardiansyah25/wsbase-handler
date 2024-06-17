package winnet

const (
	/**
	Translation: Cannot send after socket shutdown.
	Description: A request to send or receive data was not permitted because the socket had already
	been shut down in that direction with a previous shutdown (Wsapiref_60z6.asp) call.
	When a shutdown is called, a partial close of a socket is requested.
	This is a signal that the sending or receiving processes (or both) have been discontinued.
	*/
	WSAESHUTDOWN = 10058

	/*
		Translation: Connection timed out.
		Description: A connection attempt failed because the connected party did not correctly respond after
		a period of time, or the established connection failed because the connected host failed to respond.
	*/
	WSAETIMEDOUT = 10060

	/*
		Translation: Connection refused.
		Description: No connection can be made because the destination computer actively refuses it.
		This error typically results from trying to connect to a service that is inactive on the foreign host,
		that is, one that does not have a server program running.
	*/
	WSAECONNREFUSED = 10061

	/*
		Translation: Host is down.
		Description: A socket operation failed because the destination host is down.
		A socket operation encountered a dead host.
		Networking activity on the local host has not been initiated.
		These conditions are more likely to be indicated by the error WSAETIMEDOUT.
	*/
	WSAEHOSTDOWN = 10064

	/*
		Translation: No route to host.
		Description: A socket operation was tried to an unreachable host. See WSAENETUNREACH.
	*/
	WSAEHOSTUNREACH = 10065

	/*
		Translation: Network dropped connection on reset.
		Description: The connection has been broken because of keep-alive activity that detects a
		failure while the operation was in progress.
		It can also be returned by setsockopt (Wsapiref_94aa.asp) if an attempt is made to set
		SO_KEEPALIVE on a connection that has already failed.
	*/
	//WSAENETRESET = 10052

	/*
		Translation: Software caused connection abort.
		Description: An established connection was stopped by the software in your host computer,
		possibly because of a data transmission time-out or protocol error.
	*/
	WSAECONNABORTED = 10053

	/*
		Translation: Connection reset by peer
		Description: An existing connection was forcibly closed by the remote host.
		This error typically occurs if the peer program on the remote host is suddenly stopped,
		the host is restarted, or the remote host uses a hard close.
		See setsockopt (Wsapiref_94aa.asp) for more information about the SO_LINGER option on
		the remote socket.
		This error may also result if a connection was broken because of keep-alive activity that
		detects a failure while one or more operations are in progress.
		Operations that were in progress fail with WSAENETRESET.
		Subsequent operations fail with WSAECONNRESET.
	*/
	WSAECONNRESET = 10054
)
