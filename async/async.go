// Package async provides a mechanism for running asynchronous tasks
// outside of an HTTP handler for example.
//
// Async is useful for returning from an HTTP handler even if the work isn't
// able to start right away due to the channel being full.
// Implementation is based off of a post by Marcio Castilho
// found here: http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/
//
// Additionally, async provides the Invoker interface to invoke special types
// of functions. Currently lambda functions are supported, but this could also
// push messages to a more durable message queue for example.
package async
