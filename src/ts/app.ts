import { MapBoxElement } from "@/custom-elements/mapbox";

registerCustomElement("mapbox-element", MapBoxElement);

function registerCustomElement(
  name: string,
  constructor: CustomElementConstructor,
) {
  if (customElements.get(name) === undefined) {
    customElements.define(name, constructor);
  }
}
