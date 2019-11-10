package main

// TinyGo version of the 1st WebGL Fundamentals lesson
// https://webglfundamentals.org/webgl/lessons/webgl-fundamentals.html

import (
	"math/rand"
	"syscall/js"

	"github.com/justinclift/webgl"
)

var (
	// Vertex shader source code
	vertCode = `
	// an attribute will receive data from a buffer
	attribute vec2 a_position;

	uniform vec2 u_resolution;

	// all shaders have a main function
	void main() {
		// convert the position from pixels to 0.0 to 1.0
		vec2 zeroToOne = a_position.xy / u_resolution;
		
		// convert from 0->1 to 0->2
		vec2 zeroToTwo = zeroToOne * 2.0;
		
		// convert from 0->2 to -1->+1 (clip space)
		vec2 clipSpace = zeroToTwo - 1.0;
		
		gl_Position = vec4(clipSpace * vec2(1, -1), 0, 1);
	}`

	// Fragment shader source code
	fragCode = `
	precision mediump float;
	
	uniform vec4 u_color;

	void main() {
		gl_FragColor = u_color;
	}`
)

func main() {
	// Set up the WebGL context
	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", "mycanvas")
	width := canvas.Get("clientWidth").Int()
	height := canvas.Get("clientHeight").Int()
	canvas.Call("setAttribute", "width", width)
	canvas.Call("setAttribute", "height", height)
	attrs := webgl.DefaultAttributes()
	attrs.Alpha = false
	gl, err := webgl.NewContext(&canvas, attrs)
	if err != nil {
		js.Global().Call("alert", "Error: "+err.Error())
		return
	}

	// * WebGL initialisation code *

	// Create GLSL shaders, upload the GLSL source, compile the shaders
	vertexShader := createShader(gl, webgl.VERTEX_SHADER, vertCode)
	fragmentShader := createShader(gl, webgl.FRAGMENT_SHADER, fragCode)

	// Link the two shaders into a program
	program := createProgram(gl, vertexShader, fragmentShader)

	// Look up where the vertex data needs to go
	positionAttributeLocation := gl.GetAttribLocation(program, "a_position")

	// Look up uniform locations
	resolutionUniformLocation := gl.GetUniformLocation(program, "u_resolution")

	colorUniformLocation := gl.GetUniformLocation(program, "u_color")

	// Create a buffer and put three 2d clip space points in it
	positionBuffer := gl.CreateArrayBuffer()

	// Bind it to ARRAY_BUFFER (think of it as ARRAY_BUFFER = positionBuffer)
	gl.BindBuffer(webgl.ARRAY_BUFFER, positionBuffer)

	// * WebGL rendering code *

	// Tell WebGL how to convert from clip space to pixels
	gl.Viewport(0, 0, width, height)

	// Clear the canvas
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(webgl.COLOR_BUFFER_BIT)

	// Tell it to use our program (pair of shaders)
	gl.UseProgram(program)

	// Turn on the attribute
	gl.EnableVertexAttribArray(positionAttributeLocation)

	// Bind the position buffer
	gl.BindBuffer(webgl.ARRAY_BUFFER, positionBuffer)

	// Tell the attribute how to get data out of positionBuffer (ARRAY_BUFFER)
	pbSize := 2           // 2 components per iteration
	pbType := webgl.FLOAT // the data is 32bit floats
	pbNormalize := false  // don't normalize the data
	pbStride := 0         // 0 = move forward size * sizeof(pbType) each iteration to get the next position
	pbOffset := 0         // start at the beginning of the buffer
	gl.VertexAttribPointer(positionAttributeLocation, pbSize, pbType, pbNormalize, pbStride, pbOffset)

	// Set the resolution
	gl.Uniform2f(resolutionUniformLocation, float32(width), float32(height))

	// Draw 50 random rectangles in random colors
	for i := 0; i < 50; i++ {
		// Setup a random rectangle
		// This will write to positionBuffer because
		// its the last thing we bound on the ARRAY_BUFFER
		// bind point
		setRectangle(gl, float32(rand.Intn(300)), float32(rand.Intn(300)), float32(rand.Intn(300)), float32(rand.Intn(300)))

		// Set a random color
		gl.Uniform4f(colorUniformLocation, rand.Float32(), rand.Float32(), rand.Float32(), 1)

		// Draw the rectangle
		primType := webgl.TRIANGLES
		primOffset := 0
		primCount := 6
		gl.DrawArrays(primType, primOffset, primCount)
	}
}

func createShader(gl *webgl.Context, shaderType int, source string) *js.Value {
	shader := gl.CreateShader(shaderType)
	gl.ShaderSource(shader, source)
	gl.CompileShader(shader)
	success := gl.GetShaderParameter(shader, webgl.COMPILE_STATUS).Bool()
	if success {
		return shader
	}
	println(gl.GetShaderInfoLog(shader))
	gl.DeleteShader(shader)
	return &js.Value{}
}

func createProgram(gl *webgl.Context, vertexShader *js.Value, fragmentShader *js.Value) *js.Value {
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)
	success := gl.GetProgramParameterb(program, webgl.LINK_STATUS)
	if success {
		return program
	}
	println(gl.GetProgramInfoLog(program))
	gl.DeleteProgram(program)
	return &js.Value{}
}

// Fill the buffer with the values that define a rectangle
func setRectangle(gl *webgl.Context, x, y, width, height float32) {
	x1 := x
	x2 := x + width
	y1 := y
	y2 := y + height
	positionsNative := []float32{
		x1, y1,
		x2, y1,
		x1, y2,
		x1, y2,
		x2, y1,
		x2, y2,
	}
	positions := webgl.SliceToTypedArray(positionsNative)
	gl.BufferData(webgl.ARRAY_BUFFER, positions, webgl.STATIC_DRAW)
}
