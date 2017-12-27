#version 330 core

in vec3 vertex;
in vec3 color;

uniform mat4 mvMatrix;
uniform mat4 pMatrix;

out vec3 col;
void main() {
	col = color;
	gl_Position = pMatrix * mvMatrix * vec4(vertex, 1.0);
}
