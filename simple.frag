#version 330 core

in vec3 color;
out vec4 fragment;

void main() {
	fragment = vec4(color.rgb, 1.0);
}
