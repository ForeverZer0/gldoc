# GLdoc

GLdoc hosts a local server that returns OpenGL documentation parsed from the [OpenGL-Refpages](https://github.com/KhronosGroup/OpenGL-Refpages) in a simplified JSON interface, which can then be read and used from any language using only basic HTTP requests.

## Purpose

Generating bindings for OpenGL using the [OpenGL registry](https://github.com/KhronosGroup/OpenGL-Registry) is standard boilerplate for many projects. While generating the API bindings is typically trivial, including documentation for said bindings is often less so. Beyond being not required for functionality, the documentation sources are primarily intended for generating a static website in HTML, and less so for code documentation, which makes including it often not worth the extra effort. As someone who likes well-documented code and is an avid enjoyer of inline hints, this tool was created to solve that (admittedly niche) problem.

## How To Use

### Start the Server
```bash
gldoc --api=gl --version=3.3
```
There are few very options for the server, the defaults will likely be suitable for most use-cases.
```
Usage:
  gldoc [flags]

Flags:
      --api string        target OpenGL API ("gl" or "gles") (default "gl")
  -h, --help              help for gldoc
      --host string       address the server will handle requests on (default "localhost")
      --port int          port the server will handle requests on (default 8888)
      --version float32   target version for the OpenGL API, or 0 for any/latest
```

### Make Requests
Send GET requests to the server using OpenGL function names as the endpoints, whether it be from a CLI tool, a browser, or HTTP client in a programming language of your choice.

This example simply uses `curl` to make a request for `glBufferData`, then pipes the output to `jq` to unminify and format it for clarity...
```bash
curl --silent localhost:8888/glBufferData | jq
```

### Process Results

...and the resulting reply from above:
```json
{
  "name": "glBufferData",
  "desc": "creates and initializes a buffer object's data store",
  "args": [
    {
      "name": "target",
      "desc": "Specifies the target to which the buffer object is bound for glBufferData, which must be one of the buffer binding targets in the following table:"
    },
    {
      "name": "size",
      "desc": "Specifies the size in bytes of the buffer object's new data store."
    },
    {
      "name": "data",
      "desc": "Specifies a pointer to data that will be copied into the data store for initialization, or NULL if no data is to be copied."
    },
    {
      "name": "usage",
      "desc": "Specifies the expected usage pattern of the data store. The symbolic constant must be GL_STREAM_DRAW, GL_STREAM_READ, GL_STREAM_COPY, GL_STATIC_DRAW, GL_STATIC_READ, GL_STATIC_COPY, GL_DYNAMIC_DRAW, GL_DYNAMIC_READ, or GL_DYNAMIC_COPY."
    }
  ],
  "seealso": [
    "glBindBuffer",
    "glBufferSubData",
    "glMapBuffer",
    "glUnmapBuffer"
  ],
  "errors": [
    "GL_INVALID_ENUM",
    "GL_INVALID_VALUE",
    "GL_INVALID_OPERATION",
    "GL_OUT_OF_MEMORY"
  ]
}
```

This can be easily read and used to create inline documentation when generating OpenGL bindings, or whatever purpose you may need it for, with very little extra effort.