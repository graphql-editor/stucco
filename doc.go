/*

Stucco is a http server for GraphQL.

It parses server's GraphQL SDL along with it's own configuration and runs a http service ready to run GraphQL queries.

All functionality is defined in a serverless-like fashion, dispatching an actual field/type/parse/serialize functions to user functions.

Field resolution, if defined, is handled by user defined function. Otherwise the default behaviour is to return a matching field value from parent object.

Interface/Union type matching, if defined, is handled by user defined function. Otherwise it matches a type by a value of __typename field in a data. It is an error if there's no user function defined or an object is missing __typename field.

Scalar parse/serialize is handled by user defined function. If they are not defined, data is returned as is without modification.

Usage:
  -add_dir_header
    	If true, adds the file directory to the header
  -alsologtostderr
    	log to standard error as well as files
  -log_backtrace_at value
    	when logging hits line file:N, emit a stack trace
  -log_dir string
    	If non-empty, write log files in this directory
  -log_file string
    	If non-empty, use this log file
  -log_file_max_size uint
    	Defines the maximum size a log file can grow to. Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
  -logtostderr
    	log to standard error instead of files (default true)
  -skip_headers
    	If true, avoid header prefixes in the log messages
  -skip_log_headers
    	If true, avoid headers when opening log files
  -stderrthreshold value
    	logs at or above this threshold go to stderr (default 2)
  -v value
    	number for the log level verbosity (default 3)
  -version
    	print version and exit
  -vmodule value
    	comma-separated list of pattern=N settings for file-filtered logging

*/
package main
