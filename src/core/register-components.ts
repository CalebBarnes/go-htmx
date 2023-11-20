import { Component } from "@/core/component";

export type Components = {
  [componentName: string]: (props: any, render: () => void) => void;
};

export function registerComponents(components: Components) {
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
