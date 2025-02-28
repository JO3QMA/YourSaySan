# frozen_string_literal: true

require 'rubygems'
require 'bundler/setup'
require 'config'
require 'discordrb'
require 'pathname'

require_relative './voicevox'

# Discord Botのモジュール
module YouSaySan
  CONFIG = Config.load_and_set_settings('./config.yml', './command.yml')
  TOKEN = CONFIG.bot.token
  CLIENT_ID = CONFIG.bot.client_id
  PREFIX = CONFIG.bot.prefix
  MODULE_PATH = CONFIG.bot.module_path || 'core/modules'

  Discordrb::LOGGER.streams << File.open('bot2.log', 'a')
  # Discordrb::LOGGER.mode = :debug
  BOT = Discordrb::Commands::CommandBot.new(
    token: TOKEN,
    client_id: CLIENT_ID,
    prefix: PREFIX,
    ignore_bots: true
  )
  
  @text_channels = []
  @voicevox = VoiceVox.new(CONFIG, Logger.new('bot.log', 'daily').tap { |l| l.datetime_format = '%Y-%m-%d %H:%M:%S' })

  class << self
    attr_accessor :text_channels, :voicevox
  end

  # Module Loader
  def self.load_module(cls, path)
    puts "Init実行開始: #{cls}"
    new_module = Module.new
    const_set(cls.to_sym, new_module)
    Dir["#{MODULE_PATH}/#{path}/*.rb"].each do |file|
      puts "Load module : #{file}"
      load file
    rescue StandardError => e
      puts "Error loading module #{file}: #{e.message}"
    end
    
    # EventContainerとCommandContainerを分ける
    new_module.constants.each do |mod|
      mod_instance = new_module.const_get(mod)
      if mod_instance.is_a?(Class) && mod_instance.ancestors.include?(Discordrb::Commands::CommandContainer)
        BOT.include!(mod_instance)
      elsif mod_instance.is_a?(Module) && mod_instance.ancestors.include?(Discordrb::EventContainer)
        BOT.include!(mod_instance)
      end
    end
  end

  def self.init_bot
    # Load Module
    load_module(:Commands, 'commands')
    load_module(:Events, 'events')
  end

  def self.run
    puts 'RUN実行された'
    init_bot
    puts 'init終わった'
    
    # Bot Init
    BOT.ready do
      puts 'Bot is ready'
      BOT.game = CONFIG.bot.status
    end
      
    BOT.run
    puts 'bot実行された'
  end
end
