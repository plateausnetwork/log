/*
Package log provides an easy way to register application logs from any
other part of the application.

In the default configuration, the logging happens both on the stdout and
on disk (with log rotation already included).

Setting up the global logger

To set up the global logger, use the Setup function. This function sets up
both the file logging and the stdout logging, and let the logging facilities
available through the functions Debug, Info, Warn, and Error.

Once the setup is done, subsequent calls to Setup will be ignored. To dispose
of the current logger, use the TearDown function.

Structured logging basics

The functions Debug, Info and Warn all accept only a string as parameter, and
that is fine as long as the message to be logged is simple:

	log.Info("Server started")

As soon as you need to add more dynamic information to the log entries, you might
be tempted to do something like that:

	// BAD! Don't do that!!!
	log.Info(fmt.Sprintf("Server started on port %d", port))

	// will look like this on the logfile:
	// {"level":"info","msg":"Server started on port 8080", ...}

While this works on the surface, this approach is problematic because an important
piece of information (in this example, the server port) is now 'buried' into a
formatted string and logging tools cannot easily identify this.
To better handle this situation, use the With function to add custom fields as extra
logging information:

	// GOOD: structure the log with relevant information
	log.With(log.F{"port": 8080}).Info("Server started")

	// will look like this on the logfile:
	// {"level":"info","msg":"Server started", "port":8080, ...}

Note how the relevant information is now more easily accessible, both for human and machine
consumption. This is what is called 'structured logging': giving log messages clearer
and easy to parse contextual information.

The With function generates a log entry (i.e. it does not print anything) with one or more
custom fields, where the keys are strings and the values can be any arbitrary type. The log.F
type is just a handy shortcut to a map[string]interface{}. Since log entries are not printed
immediately, you need to eventually 'finalize' and entry (e.g. calling Info, Warn, etc.) to make it
show. On the other hand, the option to construct a log entry in small steps can be handy in some
situations:

	e := log.With(log.F{"port", port, "host": host})

	if customProtocol != "" {
		e = log.With(log.F{"protocol": customProtocol})
	}

	e.Info("Server started")

While this example is anecdotal, it is important to know that the possibility of exposing more (or less)
information on a log entry exists.

Logging errors

While most of the log functions and entry methods accept a string as argument, the function Error
accepts an error instance instead. This function extracts the error message and the stack trace of the
error (more on stack trace extraction below) and prints the relevant log entry:

	err := errors.New("some error")
	log.Error(err)

	// will output as
	// {"level":"error","msg":"some error", "stack": ...}

If you want to add a custom message or custom fields to and error entry, use a combination of With and
WithError calls:

	// Using a custom message
	log.WithError(err).Error("oops!")
	//output: {"level":"error", "msg":"oops", "error":"some error", "stack": ...}

	// Using custom fields
	log.With(log.F{"port": 8080}).Error(err)
	//output: {"level":"error", "msg":"some error", "port":8080, "stack": ...}

	// You can have it both ways too (both examples give the same result)
	log.With(log.F{"port": 8080}).WithError(err).Error("oops")
	log.WithError(err).With(log.F{"port": 8080}).Error("oops")
	//output: {"level":"error", "msg":"oops", "port":8080, "error":"some error", "stack": ...}

Note that once an error is added to an entry (using WithError) the entry itself locked into
an 'error state', that introduces two big changes: (1) it is not possible to use the non-error
methods to finalize the entry, and (2) the regular Error(error) method signature is changed to
Error(string), to easily accommodate a complementary error message. In other words:

	// does not compile: Debug, Info and Warn methods don't exist in this context
	log.WithError(err).Info("something")

	// does not compile: Errors expect a string here
	log.WithError(err).Error(errors.New("some error"))

	// ok - provide a custom message
	log.WithError(err).Error("something")

	// ok - no custom message (identical to log.Error(err))
	log.WithError(err).Error("")

Stack trace information

First things first: the stack trace is only visible in the log files, never on screen output.

When a log entry in the error level is created, the entry records not only the error
message, but the existing stack trace too. If using the pkg/errors package (https://github.com/pkg/errors),
the stack trace is embedded in the error instance at the creation.

However, when using the error package of the stdlib, no stack trace is created. In these cases, a
brand new stack trace is created from the point the error is logged. While useful, this may not point
to the exact root cause of the problem:

	func errorFunc() error {
		...
		return errors.New("error here!") // from stdlib
	}

	func someFunc() error {
		...
		if err := errorFunc(); err != nil {
			return err
		}

		return nil
	}

	func myFunc() {
		...
		if err := someFunc(); err != nil {
			log.Error(err)
			return
		}
	}

In the example above example, the stack trace will register the 'origin'
of the error in myFunc as oposed to the real location of the problem,
errorFunc. This is expected, because since the stdlib does not generate
a trace, the logging library generates one from scratch, in the point that
the log happens. If the 'correct' place of the error is crucial, consider using
the pkg/errors constructor to create the error.

Note that while the logger recognizes the pkg/errors generated stack, it DOES NOT
recognize the error Cause or other library-specific information.
*/
package log
