/**
 * Saved components (the `Component` GraphQL type): named HTML boxes that can
 * be placed on any canvas page as a copy or linked — full CRUD, built on
 * createCrudStore(), plus the box-side actions.
 *
 * A linked box shows its component's html, so `update` is wrapped to push a
 * new html into every loaded page's boxes (stores/boxes.js patchLinked), and
 * `remove` unlinks them locally the way the server just did (each keeps the
 * html it showed). `uses` is kept roughly right locally (countUse) as this
 * tab places, detaches and deletes linked boxes, and the panel refetches the
 * list when it opens for the true figure.
 *
 * Admin only, so nothing fetches it until an admin opens the canvas's
 * Components panel or "+ New box" menu (views/home3/ComponentPanel.vue,
 * BoxCanvas.vue).
 */

import { createCrudStore } from "@/stores/createCrudStore";
import { gql } from "@/graphql";
import { patchLinked, useBoxesStore } from "@/stores/boxes";
import { TEMPLATES } from "@/views/home3/templates";

const COMPONENT_FIELDS = `
  id
  name
  title
  html
  w
  h
  framed
  shadow
  adminOnly
  uses
`;

// Matches BOX_FIELDS in stores/boxes.js, for the mutations that answer with
// a box.
const BOX_FIELDS = `
  id page kind widget html title x y w h z framed shadow hideOnMobile adminOnly componentId
`;

export const useComponentsStore = createCrudStore("components", {
  type: "Component",
  list: "components",
  fields: COMPONENT_FIELDS,
  // Created ones are put back in name order by sort() below.
  insert: "append",
  deleteReturns: "object",
  extend: ({ items, fetch, create, update, remove }) => {
    const sort = () => items.value.sort((a, b) => a.name.localeCompare(b.name));
    const countUse = (id, delta) => {
      const c = items.value.find((c) => c.id === id);
      if (c) c.uses = Math.max(0, c.uses + delta);
    };

    return {
      countUse,

      async create(input) {
        const created = await create(input);
        sort();
        return created;
      },

      async update(id, input) {
        const updated = await update(id, input);
        if ("html" in input) patchLinked(id, { html: updated.html });
        sort();
        return updated;
      },

      async remove(id) {
        await remove(id);
        patchLinked(id, { componentId: null });
      },

      /**
       * Make html box `box` into a component named `name` and link it.
       * Throws on failure.
       */
      async saveFromBox(box, name) {
        const data = await gql(
          `mutation SaveBoxAsComponent($id: ID!, $name: String!) {
            saveBoxAsComponent(id: $id, name: $name) { ${BOX_FIELDS} }
          }`,
          { id: box.id, name },
        );
        useBoxesStore(box.page).replaceLocal(data.saveBoxAsComponent);
        await fetch();
      },

      /** Give linked box `box` its own copy of the html. Throws on failure. */
      async detach(box) {
        const data = await gql(
          `mutation DetachBox($id: ID!) { detachBox(id: $id) { ${BOX_FIELDS} } }`,
          { id: box.id },
        );
        useBoxesStore(box.page).replaceLocal(data.detachBox);
        countUse(box.componentId, -1);
      },

      /**
       * Put `component` on page key `page` at `at` ({ x, y, z }): linked, or
       * as an ordinary box with a copy of its html. Answers with the new box.
       * Throws on failure.
       */
      async place(component, page, at, linked) {
        const input = {
          kind: "html",
          title: component.title,
          w: component.w,
          h: component.h,
          framed: component.framed,
          shadow: component.shadow,
          adminOnly: component.adminOnly,
          ...at,
        };
        if (linked) input.componentId = component.id;
        else input.html = component.html;
        const created = await useBoxesStore(page).create(input);
        if (linked) countUse(component.id, 1);
        return created;
      },

      /**
       * Create a component from each built-in starter in templates.js whose
       * name is free, and answer with how many were added. Deleted ones come
       * back; ones still there are left as they are. Throws on failure.
       */
      async importBuiltins() {
        const taken = new Set(items.value.map((c) => c.name));
        let added = 0;
        for (const t of TEMPLATES) {
          if (taken.has(t.label)) continue;
          await create({
            name: t.label,
            title: t.title,
            html: t.html,
            w: t.size[0],
            h: t.size[1],
            framed: t.framed ?? true,
            adminOnly: t.adminOnly ?? false,
          });
          added++;
        }
        sort();
        return added;
      },
    };
  },
});
