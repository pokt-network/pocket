# Feature: Upgrade Protocol Cancel

#   Scenario: User Successfully Cancels a Scheduled Upgrade using CLI
#     Given the user is an ACL Owner
#     And the specified upgrade is scheduled and not yet activated
#     When the user runs the command with no error "gov cancel_upgrade"
#     Then the system should cancel the scheduled upgrade
#     When user runs the command "gov query upgrade"
#     Then the system should return the successful cancellation status

#   Scenario: User Attempts to Cancel a Past Upgrade using CLI
#     Given the user is an ACL Owner
#     And the user has a cancel upgrade command for a past version
#     When the user runs the command with no error "gov cancel_upgrade"
#     Then the system should validate the command
#     And the system should reject the command as it cannot cancel a past upgrade
