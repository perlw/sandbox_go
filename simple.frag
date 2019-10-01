#version 330 core

in vec3 fcolor;
out vec4 fragment;

void main() {
	fragment = vec4(fcolor.rgb, 1.0);
}
