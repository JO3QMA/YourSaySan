# frozen_string_literal: true

require 'rubygems'
require 'bundler/setup'
require 'config'
require 'discordrb'

require_relative './voicevox'

# Discord Botのモジュール
module YouSaySan
  CONFIG = Config.load_and_set_settings('./config.yml', './command.yml')
  TOKEN = CONFIG.bot.token
  CLIENT_ID = CONFIG.bot.client_id
  PREFIX = CONFIG.bot.prefix
  Discordrb::LOGGER.streams << File.open('bot2.log', 'a')
  # Discordrb::LOGGER.mode = :debug
  BOT = Discordrb::Commands::CommandBot.new(
    token: TOKEN,
    client_id: CLIENT_ID,
    prefix: PREFIX,
    ignore_bots: true
  )

  # Const
  @text_channels = []

  # Module Loader
  def self.load_module(cls, path)
    new_module = Module.new
    const_set(cls.to_sym, new_module)
    Dir["core/modules/#{path}/*.rb"].each do |file|
      puts "Load module : #{file}"
      load file
    end
    new_module.constants.each do |mod|
      BOT.include!(new_module.const_get(mod))
    end
  end

  # Load Module
  load_module(:Commands, 'commands')
  load_module(:Events, 'events')

  BOT.run
end
