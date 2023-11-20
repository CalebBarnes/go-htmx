import { registerComponents } from "@/core/register-components";

import {
  ExampleTimerComponent,
  ExampleComponent,
} from "@/custom-elements/example-component";
import { MapboxComponent } from "@/custom-elements/mapbox";

registerComponents({
  "mapbox-component": MapboxComponent,
  "example-component": ExampleComponent,
  "example-timer": ExampleTimerComponent,
});

// document.querySelector("#my-test-form")?.addEventListener("submit", (event) => {
//   event.preventDefault();
//   const form = event.target as HTMLFormElement;
//   const formData = new FormData(form);

//   for (const [key, value] of formData.entries()) {
//     console.log(`${key}: ${value}`);
//   }
//   // const data = Object.fromEntries(formData.entries());
//   // console.log(data);
// });
