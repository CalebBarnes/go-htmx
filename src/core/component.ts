type ComponentProps = {
  [key: string]: any;
};

export class Component extends HTMLElement {
  constructor() {
    super();
    this.props = this.getProps();
  }

  props: { [key: string]: any };

  getProps() {
    const attributes = this.attributes;
    const props: ComponentProps = {};

    for (let i = 0; i < attributes.length; i++) {
      const attribute = attributes[i];
      const key = attribute.name;

      const camelCasedName = key.replace(/-([a-z])/g, (g) =>
        g[1].toUpperCase(),
      );
      if (attribute.value === "true") {
        props[camelCasedName] = true;
      } else if (attribute.value === "false") {
        props[camelCasedName] = false;
      } else if (attribute.value === "") {
        props[camelCasedName] = true;
      } else if (attribute.value.startsWith("{")) {
        props[camelCasedName] = JSON.parse(attribute.value);
      } else if (attribute.value.startsWith("[")) {
        props[camelCasedName] = JSON.parse(attribute.value);
      } else if (!isNaN(parseFloat(attribute.value))) {
        props[camelCasedName] = Number(attribute.value);
      } else {
        props[camelCasedName] = attribute.value;
      }
    }

    return props;
  }

  public render(str: string) {
    this.innerHTML = str;
  }

  connectedCallback() {
    if (this.mount) {
      const result = this.mount();
      if (result) {
        this.render(result);
      }
    }
  }
}
