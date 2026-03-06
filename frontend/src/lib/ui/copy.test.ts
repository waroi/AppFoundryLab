import { describe, expect, it } from "vitest";
import { booleanLabel } from "./copy";

describe("booleanLabel", () => {
  describe("locale: en", () => {
    it("returns enabled/disabled for enabledDisabled mode", () => {
      expect(booleanLabel("en", true, "enabledDisabled")).toBe("enabled");
      expect(booleanLabel("en", false, "enabledDisabled")).toBe("disabled");
    });

    it("returns on/off for onOff mode", () => {
      expect(booleanLabel("en", true, "onOff")).toBe("on");
      expect(booleanLabel("en", false, "onOff")).toBe("off");
    });

    it("returns required/optional for requiredOptional mode", () => {
      expect(booleanLabel("en", true, "requiredOptional")).toBe("required");
      expect(booleanLabel("en", false, "requiredOptional")).toBe("optional");
    });

    it("returns yes/no for yesNo mode", () => {
      expect(booleanLabel("en", true, "yesNo")).toBe("yes");
      expect(booleanLabel("en", false, "yesNo")).toBe("no");
    });
  });

  describe("locale: tr", () => {
    it("returns etkin/devre disi for enabledDisabled mode", () => {
      expect(booleanLabel("tr", true, "enabledDisabled")).toBe("etkin");
      expect(booleanLabel("tr", false, "enabledDisabled")).toBe("devre disi");
    });

    it("returns acik/kapali for onOff mode", () => {
      expect(booleanLabel("tr", true, "onOff")).toBe("acik");
      expect(booleanLabel("tr", false, "onOff")).toBe("kapali");
    });

    it("returns zorunlu/opsiyonel for requiredOptional mode", () => {
      expect(booleanLabel("tr", true, "requiredOptional")).toBe("zorunlu");
      expect(booleanLabel("tr", false, "requiredOptional")).toBe("opsiyonel");
    });

    it("returns evet/hayir for yesNo mode", () => {
      expect(booleanLabel("tr", true, "yesNo")).toBe("evet");
      expect(booleanLabel("tr", false, "yesNo")).toBe("hayir");
    });
  });
});
