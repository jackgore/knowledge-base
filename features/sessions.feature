Feature: Knowledge Base user login
  In order for users to make use of our service
  They should be able to login to their accounts

Scenario Outline: The application has a login endpoint
  Given I do have an account
  When  I am not already logged in
  And I login with <isValid> credentials
  Then I should see a <code> response to login
  And  I should a see a Set-Cookie header

  Examples:
	| isValid | code |
	| true    | 200  |
	| false   | 401  |
	
