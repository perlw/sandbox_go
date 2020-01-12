package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/perlw/sandbox_go/foo"
	"github.com/perlw/sandbox_go/pkg/fontloader"
)

func init() {
	runtime.LockOSThread()
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("Using %d CPUs\n", runtime.NumCPU())
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func calcNormal(p1, p2, p3 mgl32.Vec3) mgl32.Vec3 {
	u := p2.Sub(p1)
	v := p3.Sub(p1)
	return u.Cross(v)
}

type Mesh struct {
	verts      []mgl32.Vec3
	normals    []mgl32.Vec3
	numVerts   int
	numNormals int
}

func generateMesh(width, height int) Mesh {
	mesh := Mesh{
		verts:   []mgl32.Vec3{},
		normals: []mgl32.Vec3{},
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var rootVerts []mgl32.Vec3

			offsetx := float32(0.0)
			offsety := float32(0.0)
			if y%2 == 0 {
				origox := float32(x) * 1.1
				origoy := -(float32(y) / 2.0) * 1.1

				rootVerts = []mgl32.Vec3{
					// Top
					{origox + offsetx + 0.0, 2.0, origoy + offsety + 0.5},
					{origox + offsetx + 0.5, 2.0, origoy + offsety - 0.5},
					{origox + offsetx - 0.5, 2.0, origoy + offsety - 0.5},

					// Bottom
					{origox + 0.0, 0.0, origoy + 0.5},
					{origox + 0.5, 0.0, origoy - 0.5},
					{origox - 0.5, 0.0, origoy - 0.5},
				}
			} else {
				origox := (float32(x) + 0.5) * 1.1
				origoy := -((float32(y) / 2.0) - 0.5) * 1.1

				rootVerts = []mgl32.Vec3{
					// Top
					{origox + offsetx - 0.0, 2.0, origoy + offsety - 0.5},
					{origox + offsetx - 0.5, 2.0, origoy + offsety + 0.5},
					{origox + offsetx + 0.5, 2.0, origoy + offsety + 0.5},

					// Bottom
					{origox - 0.0, 0.0, origoy - 0.5},
					{origox - 0.5, 0.0, origoy + 0.5},
					{origox + 0.5, 0.0, origoy + 0.5},
				}
			}
			rootNormals := []mgl32.Vec3{
				// Cap
				calcNormal(rootVerts[0], rootVerts[1], rootVerts[2]),

				// Front right
				calcNormal(rootVerts[3], rootVerts[1], rootVerts[0]),

				// Front left
				calcNormal(rootVerts[3], rootVerts[0], rootVerts[2]),

				// Back
				calcNormal(rootVerts[1], rootVerts[4], rootVerts[5]),
			}

			var col mgl32.Vec3
			if x%3 == 0 {
				col = mgl32.Vec3{1.0, 0.0, 0.0}
			} else if x%3 == 1 {
				col = mgl32.Vec3{0.0, 1.0, 0.0}
			} else if x%3 == 2 {
				col = mgl32.Vec3{0.0, 0.0, 1.0}
			}
			mesh.verts = append(mesh.verts, []mgl32.Vec3{
				// +Cap
				rootVerts[0],
				col,
				rootVerts[1],
				col,
				rootVerts[2],
				col,
				// -Cap

				// +Front right
				rootVerts[3],
				col,
				rootVerts[1],
				col,
				rootVerts[0],
				col,

				rootVerts[3],
				col,
				rootVerts[4],
				col,
				rootVerts[1],
				col,
				// -Front right

				// +Front left
				rootVerts[3],
				col,
				rootVerts[0],
				col,
				rootVerts[2],
				col,

				rootVerts[3],
				col,
				rootVerts[2],
				col,
				rootVerts[5],
				col,
				// -Front left

				// +Back
				rootVerts[1],
				col,
				rootVerts[4],
				col,
				rootVerts[5],
				col,

				rootVerts[1],
				col,
				rootVerts[5],
				col,
				rootVerts[2],
				col,
				// -Back
			}...)

			mesh.normals = append(mesh.normals, []mgl32.Vec3{
				// +Cap
				rootNormals[0],
				rootNormals[0],
				rootNormals[0],
				// -Cap

				// +Front right
				rootNormals[1],
				rootNormals[1],
				rootNormals[1],

				rootNormals[1],
				rootNormals[1],
				rootNormals[1],
				// -Front right

				// +Front left
				rootNormals[2],
				rootNormals[2],
				rootNormals[2],

				rootNormals[2],
				rootNormals[2],
				rootNormals[2],
				// -Front left

				// +Back
				rootNormals[3],
				rootNormals[3],
				rootNormals[3],

				rootNormals[3],
				rootNormals[3],
				rootNormals[3],
				// -Back
			}...)
		}
	}
	mesh.numVerts = len(mesh.verts)
	mesh.numNormals = len(mesh.normals)

	verts := make([]mgl32.Vec3, mesh.numVerts*3)
	copy(verts, mesh.verts)
	mesh.verts = verts

	normals := make([]mgl32.Vec3, mesh.numNormals*3)
	copy(normals, mesh.normals)
	mesh.normals = normals

	return mesh
}

func updateMesh(mesh *Mesh, width, height int, time float64, vertOffset, normalOffset int) {
	for y := 0; y < height; y++ {
		i := y * width
		for x := 0; x < width; x++ {
			// Base height
			height := 2.0 + float32((math.Sin(float64(x)+time)-math.Cos(float64(y)+time))*0.1)
			base := vertOffset + ((i + x) * 42)
			for v := 0; v < 42; v += 2 {
				if mesh.verts[base+v][1] > 0.0 {
					mesh.verts[base+v][1] = height

					// Tweaked height for wave
					mesh.verts[base+v][1] += float32(math.Sin(float64(mesh.verts[base+v][0]+mesh.verts[base+v][2])+time) * 0.25)
				}
			}
		}
	}
	for i := 0; i < mesh.numNormals; i += 21 {
		normal := calcNormal(mesh.verts[vertOffset+(i*2)+0], mesh.verts[vertOffset+(i*2)+2], mesh.verts[vertOffset+(i*2)+4])
		// Draw normal
		/*mesh.verts[(i*2)+1] = normal
		mesh.verts[(i*2)+3] = normal
		mesh.verts[(i*2)+5] = normal*/
		mesh.normals[normalOffset+i+0] = normal
		mesh.normals[normalOffset+i+1] = normal
		mesh.normals[normalOffset+i+2] = normal
	}
}

const width = 1280
const height = 720

type Glyph struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

func main() {
	foo.Foo()

	fontmap, err := fontloader.LoadTTF("pragmono.ttf")
	if err != nil {
		panic(err)
	}
	if err := fontmap.Save("out.png"); err != nil {
		panic(err)
	}

	// +Load SDF
	var sdf *image.Gray
	minBound, maxBound := 256, 0
	{
		file, err := os.Open("sdf.png")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		img, err := png.Decode(file)
		if err != nil {
			panic(err)
		}

		sdf = image.NewGray(img.Bounds())
		for y := 0; y < img.Bounds().Dy(); y++ {
			for x := 0; x < img.Bounds().Dx(); x++ {
				c := img.At(x, y).(color.Gray)
				if int(c.Y) < minBound {
					minBound = int(c.Y)
				}
				if int(c.Y) > maxBound {
					maxBound = int(c.Y)
				}
				sdf.SetGray(x, y, c)
			}
		}
	}
	// -Load SDF

	// +Setup GLFW
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.DefaultWindowHints()
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.Samples, glfw.DontCare)

	window, err := glfw.CreateWindow(width, height, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if key == glfw.KeyEscape {
			w.SetShouldClose(true)
		}
	})
	window.MakeContextCurrent()
	glfw.SwapInterval(0)
	// -Setup GLFW

	// +Setup GL
	if err := gl.Init(); err != nil {
		panic(err)
	}

	var major int32
	var minor int32
	gl.GetIntegerv(gl.MAJOR_VERSION, &major)
	gl.GetIntegerv(gl.MINOR_VERSION, &minor)
	fmt.Printf("GL %d.%d\n", major, minor)

	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.MULTISAMPLE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LESS)
	gl.Viewport(0, 0, width, height)
	gl.ClearColor(0.2, 0.2, 0.4, 1.0)
	// -Setup GL

	// +Setup font
	var fontTexture uint32
	var fontProgram uint32
	gl.GenTextures(1, &fontTexture)
	gl.BindTexture(gl.TEXTURE_2D, fontTexture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(fontmap.Image.Bounds().Dx()), int32(fontmap.Image.Bounds().Dy()), 0,
		gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(fontmap.Image.Pix))
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

	{
		vertSource, err := ioutil.ReadFile("text.vert")
		if err != nil {
			panic(err)
		}
		fragSource, err := ioutil.ReadFile("text.frag")
		if err != nil {
			panic(err)
		}
		vertSource[len(vertSource)-1] = 0
		fragSource[len(fragSource)-1] = 0
		fontProgram, err = newProgram(string(vertSource), string(fragSource))
		if err != nil {
			panic(err)
		}
	}
	gl.UseProgram(fontProgram)

	projection := mgl32.Ortho2D(0, float32(width), 0, float32(height))
	fmt.Printf("%v\n", projection)
	projectionUniform := gl.GetUniformLocation(fontProgram, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])
	/* SDF Shader
	minUniform := int32(gl.GetUniformLocation(fontProgram, gl.Str("minBound\x00")))
	gl.Uniform1i(minUniform, int32(minBound))
	maxUniform := int32(gl.GetUniformLocation(fontProgram, gl.Str("maxBound\x00")))
	gl.Uniform1i(maxUniform, int32(maxBound))
	fmt.Printf("Sending bounds %d->%d to shader\n", minBound, maxBound)
	*/

	var textVao, textVbo uint32
	gl.GenVertexArrays(1, &textVao)
	gl.BindVertexArray(textVao)
	gl.GenBuffers(1, &textVbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, textVbo)
	// sizeof(float)*6(vertices)*4(fields)
	gl.BufferData(gl.ARRAY_BUFFER, 4*6*4, nil, gl.DYNAMIC_DRAW)
	{
		vertAttrib := uint32(gl.GetAttribLocation(fontProgram, gl.Str("vertex\x00")))
		gl.EnableVertexAttribArray(vertAttrib)
		// sizeof(float)*4(fields)
		gl.VertexAttribPointer(vertAttrib, 4, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	gl.UseProgram(0)
	// -Setup font

	// +Setup shaders
	vertSource, err := ioutil.ReadFile("simple.vert")
	if err != nil {
		panic(err)
	}
	fragSource, err := ioutil.ReadFile("simple.frag")
	if err != nil {
		panic(err)
	}
	vertSource[len(vertSource)-1] = 0
	fragSource[len(fragSource)-1] = 0
	program, err := newProgram(string(vertSource), string(fragSource))
	if err != nil {
		panic(err)
	}

	gl.UseProgram(program)

	{
		projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(width)/float32(height), 1.0, 500.0)
		fmt.Printf("%v\n", projection)
		projectionUniform := gl.GetUniformLocation(program, gl.Str("pMatrix\x00"))
		gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

		model := mgl32.Translate3D(-4.0, -3.0, -10.0).Mul4(mgl32.HomogRotate3D(0.4, mgl32.Vec3{1.0, 1.5, 0.0}))
		modelUniform := gl.GetUniformLocation(program, gl.Str("mvMatrix\x00"))
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

		normal := model.Inv().Transpose()
		normalUniform := gl.GetUniformLocation(program, gl.Str("normalMatrix\x00"))
		gl.UniformMatrix4fv(normalUniform, 1, false, &normal[0])
	}
	// -Setup shaders

	// +Setup geom
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	var vbo2 uint32
	meshWidth := 128
	meshHeight := 196
	baseMesh := generateMesh(meshWidth, meshHeight)
	{
		gl.GenBuffers(1, &vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, (baseMesh.numVerts*3)*4, gl.Ptr(baseMesh.verts), gl.DYNAMIC_DRAW)

		vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertex\x00")))
		gl.EnableVertexAttribArray(vertAttrib)
		// Step (24) is not the distance between, it's the data+offset, in this case data (12) + offset (12) == 24
		gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 24, gl.PtrOffset(0))
		colorAttrib := uint32(gl.GetAttribLocation(program, gl.Str("color\x00")))
		gl.EnableVertexAttribArray(colorAttrib)
		// Offset 12 to skip the first data (which is vertex data)
		gl.VertexAttribPointer(colorAttrib, 3, gl.FLOAT, false, 24, gl.PtrOffset(12))

		gl.GenBuffers(1, &vbo2)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo2)
		gl.BufferData(gl.ARRAY_BUFFER, (baseMesh.numNormals*3)*4, gl.Ptr(baseMesh.normals), gl.DYNAMIC_DRAW)

		normalAttrib := uint32(gl.GetAttribLocation(program, gl.Str("normal\x00")))
		gl.EnableVertexAttribArray(normalAttrib)
		gl.VertexAttribPointer(normalAttrib, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	}
	gl.UseProgram(0)
	// -Setup geom

	fmt.Printf("Polys: %d | Vertices: %d | Normals: %d\n", baseMesh.numVerts/6, baseMesh.numVerts/2, baseMesh.numNormals)

	type waveMesh struct {
		mesh         *Mesh
		vertOffset   int
		normalOffset int
		timing       int
	}
	waveRebuild := make(chan waveMesh)
	go func(ch chan waveMesh) {
		tickRate := int64(1000000000) / int64(60000000)

		var vertOffset, normalOffset int
		curr := time.Now().UnixNano()
		for {
			tick := time.Now().UnixNano()
			if tick-curr >= tickRate {
				curr = tick

				timeTick := float64(tick) / 1000000000
				vertOffset += baseMesh.numVerts
				normalOffset += baseMesh.numNormals
				if vertOffset >= len(baseMesh.verts) {
					vertOffset = 0
				}
				if normalOffset >= len(baseMesh.normals) {
					normalOffset = 0
				}
				updateMesh(&baseMesh, meshWidth, meshHeight, timeTick, vertOffset, normalOffset)
				ch <- waveMesh{
					mesh:         &baseMesh,
					vertOffset:   vertOffset,
					normalOffset: normalOffset,
					timing:       int(time.Now().UnixNano()-tick) / 1000000,
				}
			}
		}
	}(waveRebuild)

	var tick float64
	var frames uint
	var fps uint
	var frameTiming, waveTiming int
	var fts, wts int
	for !window.ShouldClose() {
		frameStart := time.Now().UnixNano()

		currTick := glfw.GetTime()
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		select {
		case waveMesh := <-waveRebuild:
			waveTiming = waveMesh.timing
			gl.BindVertexArray(vao)
			gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
			// * 3 == 3 vertices, * 4 == size of float32
			gl.BufferSubData(gl.ARRAY_BUFFER, waveMesh.vertOffset, (waveMesh.mesh.numVerts*3)*4, gl.Ptr(waveMesh.mesh.verts))
			gl.BindBuffer(gl.ARRAY_BUFFER, vbo2)
			gl.BufferSubData(gl.ARRAY_BUFFER, waveMesh.normalOffset, (waveMesh.mesh.numNormals*3)*4, gl.Ptr(waveMesh.mesh.normals))
		default:
		}

		// +Draw geom
		gl.UseProgram(program)
		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(baseMesh.numVerts/2))
		// -Draw geom

		// +Render text
		type message struct {
			x, y int
			str  string
		}
		renderStrings := func(messages []message) {
			gl.UseProgram(fontProgram)
			colorUniform := int32(gl.GetUniformLocation(fontProgram, gl.Str("textColor\x00")))
			gl.Uniform3f(colorUniform, 1.0, 0.5, 0.0)
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, fontTexture)
			gl.BindVertexArray(textVao)
			vertices := make([]mgl32.Vec4, 0)
			for _, m := range messages {
				ox := float32(m.x)
				oy := float32(720 - m.y)
				stepX := float32(1.0 / 32.0)
				stepY := float32(1.0 / 16.0)
				sX := float32(8)
				sY := float32(16)
				for _, r := range m.str {
					offx := float32((r-33)%32) * stepX // - starting rune, mod 32 max characters times texture step
					offy := float32((r-33)/32) * stepY // - starting rune, div 32 max characters times texture step
					xpos := ox
					ypos := oy - sY
					vertices = append(vertices, []mgl32.Vec4{
						{xpos, ypos + sY, offx, offy},
						{xpos, ypos, offx, offy + stepY},
						{xpos + sX, ypos, offx + stepX, offy + stepY},
						{xpos, ypos + sY, offx, offy},
						{xpos + sX, ypos, offx + stepX, offy + stepY},
						{xpos + sX, ypos + sY, offx + stepX, offy},
					}...)
					ox += sX
				}
			}
			gl.BindBuffer(gl.ARRAY_BUFFER, textVbo)
			// vertices*3fields*sizeof(float)
			gl.BufferData(gl.ARRAY_BUFFER, (len(vertices)*4)*4, gl.Ptr(vertices[:]), gl.DYNAMIC_DRAW)
			gl.BindBuffer(gl.ARRAY_BUFFER, 0)
			gl.DrawArrays(gl.TRIANGLES, 0, int32(len(vertices)))
			gl.BindTexture(gl.TEXTURE_2D, 0)
		}
		messages := []message{
			{
				x: 2, y: 2,
				str: fmt.Sprintf("FPS: %d (%dms) wave timing: %dms", fps, fts, wts),
			},
			{
				x: 2, y: 20,
				str: "åäöÅÄÖ€$£#\"\\//[]{}",
			},
		}
		for i := 0; i < 20; i++ {
			messages = append(messages, message{
				x: 2, y: 38 + (i * 18),
				str: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Praesent commodo aliquam erat, quis blandit nisi interdum mollis.",
			})
		}
		renderStrings(messages)
		// -Render text

		window.SwapBuffers()
		glfw.PollEvents()

		frameTiming = int((time.Now().UnixNano() - frameStart) / 1000000)
		frames++
		if currTick-tick >= 1.0 {
			fps = frames
			fts = frameTiming
			wts = waveTiming
			frames = 0
			tick = currTick
		}
	}
}
