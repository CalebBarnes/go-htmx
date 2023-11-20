export function ExampleComponent(props: any, render: Function) {
  return /*html*/ `
  <div class="block-text border border-red-500 p-5">
    <h1>Example Function Client Component</h1>
    <h4>Hello, ${props.world}!</h4>
    <p>Example function component content</p>
    <p>${props.anotherAttribute}</p>

    <p>
    This is a function component that is registered as a custom element.
    </p>
    <h4>Props:</h4>
    <pre><code>${JSON.stringify(props, null, 2)}</code></pre>
  </div>`;
}

export function ExampleTimerComponent(props: any, render: Function) {
  let time = new Date().toLocaleTimeString();

  const renderHtml = () => {
    render(/*html*/ `
          <div class="block-text border border-red-500 p-5">
            <h1>Example Function Client Component</h1>
            <h4>Hello, ${props.world}!</h4>
            <p>Example function component content</p>
            <p>${props.anotherAttribute}</p>
            <p class="animate-fade-in">${time}</p>
            <p>
            This is a function component that is registered as a custom element.
            </p>
            <h4>Props:</h4>
            <pre><code>${JSON.stringify(props, null, 2)}</code></pre>
          </div>`);
  };

  setInterval(() => {
    time = new Date().toLocaleTimeString();
    renderHtml();
  }, 1000);

  renderHtml();
}
