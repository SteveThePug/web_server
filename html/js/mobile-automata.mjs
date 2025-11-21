function integerDigits(n, b = 10, length = null) {
  // Get the list of digits in base b
  const digits = [];
  while (n > 0) {
    digits.push(n % b);
    n = Math.floor(n / b);
  }
  digits.reverse(); // Reverse the list to get digits in big endian order

  // Pad with zeros if length is specified
  if (length !== null) {
    const padding = Array(Math.max(0, length - digits.length)).fill(0);
    return padding.concat(digits);
  }

  return digits;
}

function* cartesianProduct(...arrays) {
  // Generator for cartesian product
  if (arrays.length === 0) {
    yield [];
    return;
  }

  const [first, ...rest] = arrays;
  if (rest.length === 0) {
    for (const item of first) {
      yield [item];
    }
  } else {
    for (const item of first) {
      for (const combo of cartesianProduct(...rest)) {
        yield [item, ...combo];
      }
    }
  }
}

function tuplesFromList(lst, n) {
  const arrays = Array(n).fill(lst);
  return Array.from(cartesianProduct(...arrays));
}

function tuplesFromMultipleLists(...lists) {
  return Array.from(cartesianProduct(...lists));
}

function flattenTuples(tuples) {
  return tuples.flat();
}

function partition(lst, n) {
  const result = [];
  for (let i = 0; i < lst.length; i += n) {
    result.push(lst.slice(i, i + n));
  }
  return result;
}

function pick(pickList, lst) {
  const trues = [];
  const falses = [];
  for (let i = 0; i < pickList.length; i++) {
    if (pickList[i]) {
      trues.push(lst[i]);
    } else {
      falses.push(lst[i]);
    }
  }
  return [trues, falses];
}

function factorial(n) {
  if (n < 0) return NaN;
  if (n === 0 || n === 1) return 1;
  let result = 1;
  for (let i = 2; i <= n; i++) {
    result *= i;
  }
  return result;
}

function unrankPermutation(r, lst) {
  const n = lst.length;
  r -= 1; // Convert r to 0-indexed
  const permutation = [];
  const availableElements = [...lst];

  for (let i = n; i > 0; i--) {
    const fact = factorial(i - 1); // (n-1)!
    const index = Math.floor(r / fact); // Find the index of the current element
    permutation.push(availableElements.splice(index, 1)[0]); // Add the element and remove it from available
    r %= fact; // Update r to find the next element
  }
  return permutation;
}

export function toMaRule(sn, dn, n, k) {
  if (n < 1 || n % 2 === 0) {
    throw new Error("n must be >= 1 and odd");
  }

  const inputs = tuplesFromList([...Array(k).keys()], n);
  const directions = integerDigits(dn, 2, Math.pow(k, n)).map((x) =>
    Math.pow(-1, x),
  );
  const snDigits = integerDigits(sn, k, n * Math.pow(k, n));
  const outputs = partition(snDigits, n);

  const rules = {};
  for (let i = 0; i < inputs.length; i++) {
    rules[JSON.stringify(inputs[i])] = [outputs[i], directions[i]];
  }
  return rules;
}

export function toReversibleMaRule(bn, pn, n, k) {
  if (n < 1 || n % 2 === 0) {
    throw new Error("n must be >= 1 and odd");
  }

  const inputs = tuplesFromList([...Array(k).keys()], n);
  const blockers = tuplesFromList([...Array(k).keys()], n - 2);
  const blockSelect = pick(integerDigits(bn, 2, Math.pow(k, n - 2)), blockers);
  const rightBlockers = blockSelect[0];
  const leftBlockers = blockSelect[1];

  const twoFair = tuplesFromList([...Array(k).keys()], 2);
  const leftOutputs = tuplesFromMultipleLists(leftBlockers, twoFair).map(
    (x) => [flattenTuples(x), -1],
  );
  const rightOutputs = tuplesFromMultipleLists(twoFair, rightBlockers).map(
    (x) => [flattenTuples(x), 1],
  );

  const outputs = [...leftOutputs, ...rightOutputs];
  const rankedOutputs = unrankPermutation(pn, outputs);

  const rules = {};
  for (let i = 0; i < inputs.length; i++) {
    rules[JSON.stringify(inputs[i])] = rankedOutputs[i];
  }
  return rules;
}

export function maStep(rules, state, r) {
  /**
   * Apply one step of the mobile automaton rules
   *
   * Args:
   *   rules (object): Dictionary of rules where key is input tuple and value is [output_tuple, direction]
   *   state (array): [list, head] where list is current state and head is current position
   *   r (number): Radius of the neighborhood (window size = 2r + 1)
   *
   * Returns:
   *   array: [new_list, new_head] or [[], -1] if out of bounds
   */
  const [currentList, head] = state;

  // Check bounds
  if (head - r <= 0 || head + r >= currentList.length) {
    return [[], -1];
  }

  // Get the window of elements centered at head
  const window = currentList.slice(head - r, head + r + 1);

  // Apply rule
  const ruleKey = JSON.stringify(window);
  const [newWindow, direction] = rules[ruleKey];

  // Create new list with replaced elements
  const newList = [...currentList];
  for (let i = 0; i < newWindow.length; i++) {
    newList[head - r + i] = newWindow[i];
  }

  return [newList, head + direction];
}

export function ma(rules, initialState, t) {
  /**
   * Perform t steps of the mobile automaton
   *
   * Args:
   *   rules (object): Dictionary of rules
   *   initialState (array): Initial [list, head] state
   *   t (number): Number of steps to perform
   *
   * Returns:
   *   array: List of states at each time step
   */
  // Calculate radius from first rule key length
  const firstKey = Object.keys(rules)[0];
  const r = JSON.parse(firstKey).length / 2;

  const states = [initialState];
  let currentState = initialState;

  for (let i = 0; i < t; i++) {
    currentState = maStep(rules, currentState, r);
    states.push(currentState);

    // Stop if we hit an invalid state
    if (currentState[0].length === 0) {
      break;
    }
  }

  return states;
}

export function cyclicMaStep(rules, state, r) {
  /**
   * Cyclic version: indexing wraps around the array.
   */
  const [currentList, head] = state;
  const n = currentList.length;

  // --- Cyclic window extraction ---
  const window = [];
  for (let i = -r; i <= r; i++) {
    window.push(currentList[(head + i + n) % n]);
  }

  // Apply rule
  const ruleKey = JSON.stringify(window);
  const [newWindow, direction] = rules[ruleKey];

  // --- Cyclic writeback ---
  const newList = [...currentList];
  for (let offset = 0; offset < newWindow.length; offset++) {
    newList[(head - r + offset + n) % n] = newWindow[offset];
  }

  // Move head cyclically
  const newHead = (head + direction + n) % n;
  return [newList, newHead];
}

export function cyclicMa(rules, initialState, t) {
  /**
   * Perform t steps of the mobile automaton
   *
   * Args:
   *   rules (object): Dictionary of rules
   *   initialState (array): Initial [list, head] state
   *   t (number): Number of steps to perform
   *
   * Returns:
   *   array: List of states at each time step
   */
  // Calculate radius from first rule key length
  const firstKey = Object.keys(rules)[0];
  const r = JSON.parse(firstKey).length / 2;

  const states = [initialState];
  let currentState = initialState;

  for (let i = 0; i < t; i++) {
    currentState = cyclicMaStep(rules, currentState, r);
    states.push(currentState);

    // Stop if we hit an invalid state
    if (currentState[0].length === 0) {
      break;
    }
  }

  return states;
}
