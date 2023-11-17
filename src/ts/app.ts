import { MapBoxElement } from "@/ts/custom-elements/mapbox";

registerCustomElement("mapbox-element", MapBoxElement);

function registerCustomElement(
  name: string,
  constructor: CustomElementConstructor,
) {
  if (customElements.get(name) === undefined) {
    customElements.define(name, constructor);
  }
}

// todo: automatically add ts entry point for each page template
