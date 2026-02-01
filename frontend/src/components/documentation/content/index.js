// Documentation content - modular structure
// Each section is in its own file for maintainability

import { gettingStartedSection } from './getting-started';
import { apiReferenceSection } from './api-reference';
import { templateFormatSection } from './template-format';
import { advancedFeaturesSection } from './advanced-features';
import { examplesSection } from './examples';
import { sampleDataSection } from './sample-data';
import { pythonBindingsSection } from './python-bindings';

// Combined documentation sections in display order
export const docSections = [
    gettingStartedSection,
    apiReferenceSection,
    templateFormatSection,
    advancedFeaturesSection,
    examplesSection,
    sampleDataSection,
    pythonBindingsSection
];

// Re-export individual sections for direct access if needed
export {
    gettingStartedSection,
    apiReferenceSection,
    templateFormatSection,
    advancedFeaturesSection,
    examplesSection,

    sampleDataSection,
    pythonBindingsSection
};
