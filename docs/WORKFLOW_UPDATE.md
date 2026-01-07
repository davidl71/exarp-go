# Workflow Update: Minimize Human Review

**Date:** 2026-01-07  
**Change:** Updated Todo2 workflow to prefer marking tasks as "Done" directly, using "Review" only for critical tasks

## Summary of Changes

### ✅ New Default Workflow

**Previous:** All tasks required Review status before Done  
**New:** Tasks go directly to Done after result comment, Review only for critical tasks

### Key Changes

1. **Task Lifecycle Updated**
   - Default path: `In Progress` → `Done` (with result comment)
   - Exception path: `In Progress` → `Review` → `Done` (critical tasks only)

2. **Review Decision Framework**
   - Use Review ONLY for:
     - High-stakes changes (production deployments, security changes, breaking changes)
     - Complex implementations (major architectural changes, multi-system integrations)
     - Uncertain outcomes (experimental features, untested approaches)
     - User-requested review (explicitly requested by human)
     - High-risk operations (data migrations, critical bug fixes affecting many users)

3. **Default Behavior**
   - Routine tasks → Done directly
   - Simple changes → Done directly
   - Well-understood work → Done directly
   - **When in doubt, go directly to Done**

### Updated Sections

1. **Task Lifecycle State Machine**
   - Updated to show Done as default path
   - Review shown as exception path for critical tasks

2. **Review Workflow Section**
   - Renamed from "CRITICAL REVIEW WORKFLOW WARNING"
   - Now: "REVIEW WORKFLOW - MINIMIZE HUMAN REVIEW"
   - Added decision framework for when to use Review
   - Added examples of tasks that should go directly to Done

3. **Task Completion Requirements**
   - Updated to show both workflows (Done default, Review exception)
   - Added completion decision guidance

4. **Comment Types**
   - Updated result comment description
   - Clarified it's mandatory before Done (not Review)

5. **Enforcement Policies**
   - Added rule: "Prefer marking as Done directly (use Review only for critical tasks)"
   - Updated workflow summary

### Benefits

- ✅ **Faster workflow** - Less waiting for human approval
- ✅ **More efficient** - Routine tasks complete automatically
- ✅ **Human time saved** - Only review critical/high-stakes tasks
- ✅ **Still safe** - Critical tasks still require review
- ✅ **Flexible** - Human can still request review when needed

### Examples

**✅ Go Directly to Done:**
- Bug fix: "Fixed typo in error message"
- Documentation: "Updated README with installation steps"
- Configuration: "Added new environment variable"
- Simple feature: "Added validation for email format"
- Refactoring: "Extracted function to reduce duplication"

**🚨 Use Review:**
- Security: "Added authentication middleware"
- Breaking change: "Removed deprecated API endpoint"
- Production deployment: "Deployed new version to production"
- Data migration: "Migrated user data to new schema"
- Major architecture: "Refactored entire authentication system"

### Migration Notes

- Existing tasks in Review status will still require human approval
- New tasks will default to Done after result comment
- Humans can still request Review for any task via comments
- Critical tasks should explicitly note why Review is needed in result comment

---

**Status:** ✅ Complete - Workflow updated to minimize human review while maintaining safety for critical tasks

