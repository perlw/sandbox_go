#version 330 core
in vec2 TexCoords;
out vec4 color;

uniform sampler2D text;
uniform vec3 textColor;

void main() {
    float sampled = texture(text, TexCoords).r;
    if (sampled >= 0.513) {
        color = vec4(textColor, 1.0);
    } else {
        discard;
    }
}
