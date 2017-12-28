#version 330 core

in vec3 vertex;
in vec3 color;

uniform mat4 mvMatrix;
uniform mat4 pMatrix;

uniform float time;

out float depth;
out vec3 col;
void main() {
	col = color;

	vec3 vert = vertex;
	float x = sin((vert.x + time) * 0.5);
	vert.z += x;
	depth = (x + 1.0) / 2.0;
	gl_Position = pMatrix * mvMatrix * vec4(vert.xyz, 1.0);
}
