# frozen_string_literal: true

require 'rspec'
require 'ostruct'

RSpec.configure do |config|
	config.expect_with :rspec do |c|
		c.syntax = :expect
	end

	config.mock_with :rspec do |m|
		m.verify_partial_doubles = true
	end

	config.color = true
	config.formatter = :documentation
end
