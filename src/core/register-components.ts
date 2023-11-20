import { Component } from "@/core/component";

export function registerComponents(components: any) {
  for (const key in components) {
    const component = components[key];

    if (customElements.get(key) === undefined) {
      customElements.define(
        key,
        class extends Component {
          mount() {
            return component(this.props, this.render.bind(this));
          }
        },
      );
    }
  }
}
