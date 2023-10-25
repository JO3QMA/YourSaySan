# frozen_string_literal: true

require 'rubygems'
require 'bundler/setup'
require 'config'

# Discord Botのモジュール
module YouSaySan
  CONFIG = onfig.load_and_set_settings('./config.yml', './command.yml')
  TOKEN = CONFIG.bot.token
  CLIENT_ID = CONFIG.bot.client_id
  PREFIX = CONFIG.bot.prefix
  BOT.Discordrb::Commands::CommandBot.new(
    token: TOKEN,
    client_id: CLIENT_ID,
    prefix: PREFIX,
    ignore_bots: true
  )

  # Module Loader
  def self.load_module(cls, path)
    new_module = Module.new
    const_set(cls.to_sym, new_module)
    Dir["core/modules/#{path}/*.rb"].each do |file|
      load file
    end
    new_module.constants.each do |mod|
      BOT.include! new_module.const_get(mod)
    end
  end

  # Load Module
  load_module(:Commands, 'commands')
  load_module(:Events, 'events')

  BOT.run
end
