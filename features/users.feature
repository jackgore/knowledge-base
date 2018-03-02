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
	
Scenario Outline: The application has a get user endpoint
  Given I do have a running web server
  When I try to retrieve user with <username>
  Then I should see a <code> response

  Examples:
	| username | code |
	| notreal  | 404  |
	| real     | 200  |
	
