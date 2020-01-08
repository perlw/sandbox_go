#version 330 core
in vec2 TexCoords;
out vec4 color;

uniform sampler2D text;
uniform vec3 textColor;

void main() {
    vec4 sampled = texture(text, TexCoords);
    color = vec4(textColor, 1.0) * sampled;
}
