Feature: Knowledge Base user signup
  In order to generate revenue
  The users
  Should be able to signup for our website

Scenario Outline: The application has a signup endpoint
  Given I do have a running web server
  When I sign up with <isValid> credentials
  Then I should see a <code> response

  Examples:
	| isValid | code |
	| true    | 200  |
	| false   | 400  |
	
