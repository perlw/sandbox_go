#version 330 core

in vec3 vertex;
in vec3 normal;

uniform mat4 mvMatrix;
uniform mat4 pMatrix;
uniform mat4 normalMatrix;

out vec3 fcolor;
void main() {
	vec3 vert = vertex.xyz;

	fcolor.r = vert.x / 18.0;
	fcolor.g = vert.y / 4.0;
	fcolor.b = vert.z / 26.0;

	gl_Position = pMatrix * mvMatrix * vec4(vert.xyz, 1.0);
}
