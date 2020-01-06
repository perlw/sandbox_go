#version 330 core
in vec2 TexCoords;
out vec4 color;

uniform sampler2D text;
uniform vec3 textColor;

void main() {
    float sampled = (texture(text, TexCoords).r + texture(text, TexCoords).g + texture(text, TexCoords).b) / 3.0;
    color = vec4(textColor, sampled);
}
