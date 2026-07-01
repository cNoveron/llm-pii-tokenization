# Trustate Demo — Suggested Questions Spec

A curated set of questions surfaced in the UI as clickable suggestions.
Each maps to a real operational decision a paralegal or estate attorney
would need to make. Grouped by decision category. Rendered as chips/pills
below the chat input — clicking one populates the question input and submits.

The goal is twofold: (1) guide a non-technical demo audience toward
questions that produce impressive LLM answers, and (2) signal domain
awareness to the recruiting engineer.

---

## UI Placement

```
│  ┌─────────────────────────────────┐ [Send]     │
│  │ Ask a question…                 │            │
│  └─────────────────────────────────┘            │
│                                                  │
│  Suggested:                                      │
│  [Who are the residuary beneficiaries?]          │
│  [Are there any no-contest clauses?]             │
│  [Is there a pour-over will?]        …           │
└──────────────────────────────────────────────────┘
```

Chips are horizontally scrollable on overflow. Clicking a chip:
1. Populates `#question-input` with the question text
2. Immediately calls `askQuestion()` — no extra click needed

---

## Question Set

### Category 1 — Asset Identification & Inventory

These drive the initial estate inventory, a mandatory step before any
distribution or creditor settlement can begin.

| # | Question | Why it matters operationally |
|---|---|---|
| Q01 | What real property is included in the estate, and are there any co-ownership arrangements? | Determines whether probate applies per jurisdiction; co-ownership (joint tenancy vs tenancy in common) changes survivorship rules |
| Q02 | Are any assets held in trust and therefore outside the probate estate? | Trust assets pass outside probate — identifying them early prevents double-counting and delays |
| Q03 | Does the document reference any business interests, partnership stakes, or LLC memberships? | Business interests require valuation and may trigger buy-sell agreement reviews |
| Q04 | Are there any assets with named beneficiaries that supersede this document, such as life insurance or retirement accounts? | Beneficiary designations on accounts override will provisions — critical for accurate net estate calculation |
| Q05 | Does the estate include any foreign assets or property held in another jurisdiction? | Triggers ancillary probate proceedings in each relevant jurisdiction |

---

### Category 2 — Liability & Debt Discovery

Creditors must be satisfied before any distribution. Missing a liability
exposes the executor to personal liability.

| # | Question | Why it matters operationally |
|---|---|---|
| Q06 | Does the document acknowledge any outstanding debts, mortgages, or liens on estate property? | Liens survive the testator's death and must be cleared before transfer of title |
| Q07 | Are there any personal guarantees or co-signed obligations mentioned? | Co-signed debts may become estate liabilities depending on jurisdiction |
| Q08 | Is the executor authorized to pay debts and taxes from the estate before distribution? | Some wills restrict the executor's discretion; knowing this early prevents unauthorized disbursements |
| Q09 | Does the document reference any pending litigation or contingent liabilities? | Pending lawsuits can freeze estate assets for months or years |
| Q10 | Are there any specific bequests of encumbered property, and who bears the debt — the estate or the beneficiary? | Determines whether the estate must pay off a mortgage before transferring property, or whether the beneficiary inherits it subject to the lien |

---

### Category 3 — Beneficiary Rights & Distribution

The most contested area of estate administration. Ambiguity here causes
litigation.

| # | Question | Why it matters operationally |
|---|---|---|
| Q11 | Who are the residuary beneficiaries, and what share does each receive? | The residuary clause catches everything not covered by specific bequests — most disputes originate here |
| Q12 | Are there any conditions or contingencies attached to any bequest, such as reaching a certain age or surviving the testator by a specified period? | Conditional bequests require monitoring before disbursement — missed conditions void the gift |
| Q13 | Does the document provide for any beneficiary who is a minor, and is a guardian or custodian named for their inheritance? | Minors cannot receive assets directly — a custodian under UTMA or a testamentary trust must be established |
| Q14 | Are any beneficiaries explicitly disinherited, and does the document state a reason? | Explicit disinheritance with stated reasons is harder to contest; omission without reason is a common ground for will contests |
| Q15 | Is there a no-contest (in terrorem) clause, and which jurisdictions enforce it? | A valid no-contest clause deters litigation but is unenforceable in some US states (e.g. Florida) and irrelevant in others |

---

### Category 4 — Executor & Fiduciary Authority

Defines who has legal power to act and what they can and cannot do.

| # | Question | Why it matters operationally |
|---|---|---|
| Q16 | Who is named as executor, and is an alternate or successor executor designated? | If the primary executor predeceases or declines, the alternate takes over — without one, the court appoints an administrator |
| Q17 | What powers are granted to the executor — can they sell property, invest assets, or operate a business without court approval? | Broad independent administration authority speeds up settlement significantly; limited authority requires court petitions for every major action |
| Q18 | Is the executor entitled to compensation, and at what rate or amount? | Executor fees are taxable income and affect the net estate value available for distribution |
| Q19 | Are there any restrictions on the executor's ability to self-deal or benefit personally from the estate? | Self-dealing without explicit authorization is a breach of fiduciary duty regardless of what the will says |
| Q20 | Does the document waive the requirement for the executor to post a surety bond? | Bond waivers save estate money but expose beneficiaries to risk — relevant when evaluating executor trustworthiness |

---

### Category 5 — Trust & Testamentary Provisions

Relevant when the estate includes or pours into a trust structure.

| # | Question | Why it matters operationally |
|---|---|---|
| Q21 | Is there a pour-over will that directs assets into a revocable living trust at death? | Pour-over wills still require probate for assets not already titled to the trust — a common misconception among clients |
| Q22 | Who is named as trustee, and is there a mechanism for trustee succession? | Trustee succession gaps require court intervention to appoint a replacement |
| Q23 | Does the trust include spendthrift provisions protecting beneficiaries from creditors? | Spendthrift clauses prevent beneficiaries from assigning their interest — relevant when a beneficiary has judgments against them |
| Q24 | Are there any special needs trusts or supplemental needs provisions for a beneficiary with a disability? | Improper distribution to a beneficiary on Medicaid/SSI can disqualify them from government benefits |
| Q25 | What are the distribution standards for any discretionary trust — HEMS or purely discretionary? | HEMS is an ascertainable standard that avoids estate tax inclusion; purely discretionary grants broader trustee power but different tax treatment |

---

### Category 6 — Tax & Estate Planning Signals

Questions that surface tax exposure before a CPA or tax attorney is engaged.

| # | Question | Why it matters operationally |
|---|---|---|
| Q26 | Does the document reference a marital deduction, QTIP trust, or any provisions designed to defer estate tax? | QTIP elections must be made on the estate tax return — missing the deadline is irrecoverable |
| Q27 | Is there a credit shelter or bypass trust provision to use the deceased's estate tax exemption? | Failure to fund a credit shelter trust wastes the exemption and increases tax exposure for the surviving spouse's estate |
| Q28 | Are there any charitable bequests that qualify for an estate tax deduction? | Charitable deductions reduce the taxable estate — must be to qualified organizations under IRC §2055 |
| Q29 | Does the document reference any gifts made within three years of death that might be pulled back into the taxable estate? | Certain transfers within three years of death (e.g. life insurance policy transfers) are included in the gross estate under IRC §2035 |
| Q30 | Is there a generation-skipping transfer (GST) provision, and has a GST exemption allocation been referenced? | GST tax applies to transfers to grandchildren and below — a separate exemption allocation is required and easily missed |

---

## UI Implementation Notes

### Chip rendering (on upload success, alongside facts)

```js
const SUGGESTED_QUESTIONS = [
  "What real property is included in the estate, and are there co-ownership arrangements?",
  "Are any assets held in trust and outside the probate estate?",
  "Does the document acknowledge any outstanding debts, mortgages, or liens?",
  "Who are the residuary beneficiaries and what share does each receive?",
  "Are there any conditions or contingencies attached to any bequest?",
  "Is there a no-contest clause?",
  "What powers are granted to the executor?",
  "Is there a pour-over will that directs assets into a living trust?",
  "Are there any special needs trust provisions?",
  "Does the document reference a marital deduction or QTIP trust?",
];

function renderSuggestions() {
  suggestionsBar.innerHTML = '';
  SUGGESTED_QUESTIONS.forEach(q => {
    const chip = document.createElement('button');
    chip.className = 'suggestion-chip';
    chip.textContent = q;
    chip.addEventListener('click', () => {
      questionInput.value = q;
      askQuestion();
    });
    suggestionsBar.appendChild(chip);
  });
  suggestionsBar.hidden = false;
}
```

Call `renderSuggestions()` alongside `seedChatThread()` on upload success.

### CSS

```css
#suggestions-bar {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding: 6px 0;
  scrollbar-width: thin;
  margin-bottom: 6px;
}

#suggestions-label {
  font-size: 0.75em;
  color: #888;
  margin-bottom: 4px;
}

.suggestion-chip {
  white-space: nowrap;
  padding: 5px 12px;
  border: 1px solid #1a5276;
  border-radius: 16px;
  background: white;
  color: #1a5276;
  font-size: 0.82em;
  cursor: pointer;
  flex-shrink: 0;
}

.suggestion-chip:hover {
  background: #1a5276;
  color: white;
}
```

---

## Tasks

- [ ] **F050** — JS/HTML: `#suggestions-label` + `#suggestions-bar` container
  below `#chat-input-row`; `renderSuggestions()` called on upload success;
  chip click populates input and calls `askQuestion()`. _(deps: F045)_

- [ ] **F051** — CSS: suggestions bar styles (horizontal scroll, chips, hover).
  _(deps: F050)_

- [ ] **F052** — Verify: after loading a sample doc, suggestion chips appear;
  clicking "Is there a no-contest clause?" submits the question and appends
  the exchange to the chat thread correctly. _(deps: F051, F047)_