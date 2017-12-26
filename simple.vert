#version 330 core

in vec3 vertex;

uniform mat4 mvMatrix;
uniform mat4 pMatrix;

void main() {
    gl_Position = pMatrix * mvMatrix * vec4(vertex, 1.0);
}
