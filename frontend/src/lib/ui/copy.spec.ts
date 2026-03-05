import { describe, expect, test } from "vitest";
import { renderTemplate } from "./copy";

describe("renderTemplate", () => {
  test("replaces single key with corresponding value", () => {
    const template = "Hello {name}!";
    const values = { name: "World" };
    expect(renderTemplate(template, values)).toBe("Hello World!");
  });

  test("replaces multiple occurrences of the same key", () => {
    const template = "{word} {word} {word}";
    const values = { word: "test" };
    expect(renderTemplate(template, values)).toBe("test test test");
  });

  test("replaces multiple different keys", () => {
    const template = "{greeting} {name}, today is {day}.";
    const values = { greeting: "Hi", name: "Alice", day: "Monday" };
    expect(renderTemplate(template, values)).toBe("Hi Alice, today is Monday.");
  });

  test("replaces missing keys with empty string", () => {
    const template = "Hello {name}!";
    const values = {};
    expect(renderTemplate(template, values)).toBe("Hello !");
  });

  test("ignores extra keys in values record", () => {
    const template = "Hello {name}!";
    const values = { name: "World", extra: "ignore me" };
    expect(renderTemplate(template, values)).toBe("Hello World!");
  });

  test("works with an empty template", () => {
    const template = "";
    const values = { key: "value" };
    expect(renderTemplate(template, values)).toBe("");
  });

  test("works with template without any keys", () => {
    const template = "No keys here.";
    const values = { key: "value" };
    expect(renderTemplate(template, values)).toBe("No keys here.");
  });

  test("handles numbers as strings in values", () => {
    const template = "Count: {count}";
    const values = { count: "42" };
    expect(renderTemplate(template, values)).toBe("Count: 42");
  });
});
