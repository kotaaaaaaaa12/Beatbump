package _youtube

import (
	"encoding/json"
	"strings"
)

type HomeResponse struct {
	ResponseContext struct {
		VisitorData           string `json:"visitorData"`
		ServiceTrackingParams []struct {
			Service string `json:"service"`
			Params  []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"params"`
		} `json:"serviceTrackingParams"`
		MaxAgeSeconds int `json:"maxAgeSeconds"`
	} `json:"responseContext"`
	Contents struct {
		SingleColumnBrowseResultsRenderer struct {
			Tabs []struct {
				TabRenderer struct {
					Endpoint struct {
						ClickTrackingParams string `json:"clickTrackingParams"`
						BrowseEndpoint      struct {
							BrowseID string `json:"browseId"`
						} `json:"browseEndpoint"`
					} `json:"endpoint"`
					Title    string `json:"title"`
					Selected bool   `json:"selected"`
					Content  struct {
						SectionListRenderer struct {
							Contents      []HomeContent `json:"contents"`
							Continuations []struct {
								NextContinuationData struct {
									Continuation        string `json:"continuation"`
									ClickTrackingParams string `json:"clickTrackingParams"`
								} `json:"nextContinuationData"`
							} `json:"continuations"`
							TrackingParams string `json:"trackingParams"`
							Header         struct {
								ChipCloudRenderer struct {
									Chips []struct {
										ChipCloudChipRenderer struct {
											Style struct {
												StyleType string `json:"styleType"`
											} `json:"style"`
											Text struct {
												Runs []struct {
													Text string `json:"text"`
												} `json:"runs"`
											} `json:"text"`
											NavigationEndpoint struct {
												ClickTrackingParams string `json:"clickTrackingParams"`
												BrowseEndpoint      struct {
													BrowseID string `json:"browseId"`
													Params   string `json:"params"`
												} `json:"browseEndpoint"`
											} `json:"navigationEndpoint"`
											TrackingParams      string `json:"trackingParams"`
											IsSelected          bool   `json:"isSelected"`
											OnDeselectedCommand struct {
												ClickTrackingParams string `json:"clickTrackingParams"`
												BrowseEndpoint      struct {
													BrowseID string `json:"browseId"`
													Params   string `json:"params"`
												} `json:"browseEndpoint"`
											} `json:"onDeselectedCommand"`
										} `json:"chipCloudChipRenderer"`
									} `json:"chips"`
									TrackingParams       string `json:"trackingParams"`
									HorizontalScrollable bool   `json:"horizontalScrollable"`
								} `json:"chipCloudRenderer"`
							} `json:"header"`
						} `json:"sectionListRenderer"`
					} `json:"content"`
					Icon struct {
						IconType string `json:"iconType"`
					} `json:"icon"`
					TabIdentifier  string `json:"tabIdentifier"`
					TrackingParams string `json:"trackingParams"`
				} `json:"tabRenderer"`
			} `json:"tabs"`
		} `json:"singleColumnBrowseResultsRenderer"`
	} `json:"contents"`
	ContinuationContents struct {
		SectionListContinuation struct {
			Contents []struct {
				MusicCarouselShelfRenderer struct {
					Header struct {
						MusicCarouselShelfBasicHeaderRenderer struct {
							Title struct {
								Runs []struct {
									Text string `json:"text"`
								} `json:"runs"`
							} `json:"title"`
							Strapline struct {
								Runs []struct {
									Text string `json:"text"`
								} `json:"runs"`
							} `json:"strapline"`
							AccessibilityData struct {
								AccessibilityData struct {
									Label string `json:"label"`
								} `json:"accessibilityData"`
							} `json:"accessibilityData"`
							HeaderStyle    string `json:"headerStyle"`
							TrackingParams string `json:"trackingParams"`
						} `json:"musicCarouselShelfBasicHeaderRenderer"`
					} `json:"header"`
					Contents []struct {
						MusicResponsiveListItemRenderer *MusicResponsiveListItemRenderer `json:"musicResponsiveListItemRenderer"`
						MusicTwoRowItemRenderer         *MusicTwoRowItemRenderer         `json:"musicTwoRowItemRenderer,omitempty"`
					} `json:"contents"`
					TrackingParams string `json:"trackingParams"`
					ItemSize       string `json:"itemSize"`
				} `json:"musicCarouselShelfRenderer"`
			} `json:"contents"`
			Continuations []struct {
				NextContinuationData struct {
					Continuation        string `json:"continuation"`
					ClickTrackingParams string `json:"clickTrackingParams"`
				} `json:"nextContinuationData"`
			} `json:"continuations"`
			TrackingParams string `json:"trackingParams"`
		} `json:"sectionListContinuation"`
	} `json:"continuationContents"`
	TrackingParams     string `json:"trackingParams"`
	MaxAgeStoreSeconds int    `json:"maxAgeStoreSeconds"`
	Header             struct {
		MusicImmersiveHeaderRenderer struct {
			Title struct {
				Runs []struct {
					Text string `json:"text"`
				} `json:"runs"`
			} `json:"title"`
			SubscriptionButton struct {
				SubscribeButtonRenderer struct {
					SubscriberCountText struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
					} `json:"subscriberCountText"`
					Subscribed                       bool   `json:"subscribed"`
					Enabled                          bool   `json:"enabled"`
					Type                             string `json:"type"`
					ChannelID                        string `json:"channelId"`
					ShowPreferences                  bool   `json:"showPreferences"`
					SubscriberCountWithSubscribeText struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
					} `json:"subscriberCountWithSubscribeText"`
					SubscribedButtonText struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
					} `json:"subscribedButtonText"`
					UnsubscribedButtonText struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
					} `json:"unsubscribedButtonText"`
					TrackingParams        string `json:"trackingParams"`
					UnsubscribeButtonText struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
					} `json:"unsubscribeButtonText"`
					ServiceEndpoints []struct {
						ClickTrackingParams string `json:"clickTrackingParams"`
						SubscribeEndpoint   struct {
							ChannelIds []string `json:"channelIds"`
							Params     string   `json:"params"`
						} `json:"subscribeEndpoint,omitempty"`
						SignalServiceEndpoint struct {
							Signal  string `json:"signal"`
							Actions []struct {
								ClickTrackingParams string `json:"clickTrackingParams"`
								OpenPopupAction     struct {
									Popup struct {
										ConfirmDialogRenderer struct {
											TrackingParams string `json:"trackingParams"`
											DialogMessages []struct {
												Runs []struct {
													Text string `json:"text"`
												} `json:"runs"`
											} `json:"dialogMessages"`
											ConfirmButton struct {
												ButtonRenderer struct {
													Style string `json:"style"`
													Size  string `json:"size"`
													Text  struct {
														Runs []struct {
															Text string `json:"text"`
														} `json:"runs"`
													} `json:"text"`
													ServiceEndpoint struct {
														ClickTrackingParams string `json:"clickTrackingParams"`
														UnsubscribeEndpoint struct {
															ChannelIds []string `json:"channelIds"`
														} `json:"unsubscribeEndpoint"`
													} `json:"serviceEndpoint"`
													TrackingParams string `json:"trackingParams"`
												} `json:"buttonRenderer"`
											} `json:"confirmButton"`
											CancelButton struct {
												ButtonRenderer struct {
													Style string `json:"style"`
													Size  string `json:"size"`
													Text  struct {
														Runs []struct {
															Text string `json:"text"`
														} `json:"runs"`
													} `json:"text"`
													TrackingParams string `json:"trackingParams"`
												} `json:"buttonRenderer"`
											} `json:"cancelButton"`
										} `json:"confirmDialogRenderer"`
									} `json:"popup"`
									PopupType string `json:"popupType"`
								} `json:"openPopupAction"`
							} `json:"actions"`
						} `json:"signalServiceEndpoint,omitempty"`
					} `json:"serviceEndpoints"`
					LongSubscriberCountText struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
						Accessibility struct {
							AccessibilityData struct {
								Label string `json:"label"`
							} `json:"accessibilityData"`
						} `json:"accessibility"`
					} `json:"longSubscriberCountText"`
					ShortSubscriberCountText struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
					} `json:"shortSubscriberCountText"`
					SubscribeAccessibility struct {
						AccessibilityData struct {
							Label string `json:"label"`
						} `json:"accessibilityData"`
					} `json:"subscribeAccessibility"`
					UnsubscribeAccessibility struct {
						AccessibilityData struct {
							Label string `json:"label"`
						} `json:"accessibilityData"`
					} `json:"unsubscribeAccessibility"`
					SignInEndpoint struct {
						ClickTrackingParams string `json:"clickTrackingParams"`
						ModalEndpoint       struct {
							Modal struct {
								ModalWithTitleAndButtonRenderer struct {
									Title struct {
										Runs []struct {
											Text string `json:"text"`
										} `json:"runs"`
									} `json:"title"`
									Content struct {
										Runs []struct {
											Text string `json:"text"`
										} `json:"runs"`
									} `json:"content"`
									Button struct {
										ButtonRenderer struct {
											Style      string `json:"style"`
											IsDisabled bool   `json:"isDisabled"`
											Text       struct {
												Runs []struct {
													Text string `json:"text"`
												} `json:"runs"`
											} `json:"text"`
											NavigationEndpoint struct {
												ClickTrackingParams string `json:"clickTrackingParams"`
												SignInEndpoint      struct {
													Hack bool `json:"hack"`
												} `json:"signInEndpoint"`
											} `json:"navigationEndpoint"`
											TrackingParams string `json:"trackingParams"`
										} `json:"buttonRenderer"`
									} `json:"button"`
								} `json:"modalWithTitleAndButtonRenderer"`
							} `json:"modal"`
						} `json:"modalEndpoint"`
					} `json:"signInEndpoint"`
				} `json:"subscribeButtonRenderer"`
			} `json:"subscriptionButton"`
			Description struct {
				Runs []struct {
					Text string `json:"text"`
				} `json:"runs"`
			} `json:"description"`
			MoreButton struct {
				ToggleButtonRenderer struct {
					IsToggled   bool `json:"isToggled"`
					IsDisabled  bool `json:"isDisabled"`
					DefaultIcon struct {
						IconType string `json:"iconType"`
					} `json:"defaultIcon"`
					DefaultText struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
					} `json:"defaultText"`
					ToggledIcon struct {
						IconType string `json:"iconType"`
					} `json:"toggledIcon"`
					ToggledText struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
					} `json:"toggledText"`
					TrackingParams string `json:"trackingParams"`
				} `json:"toggleButtonRenderer"`
			} `json:"moreButton"`
			Menu struct {
				MenuRenderer struct {
					Items []struct {
						MenuNavigationItemRenderer struct {
							Text struct {
								Runs []struct {
									Text string `json:"text"`
								} `json:"runs"`
							} `json:"text"`
							Icon struct {
								IconType string `json:"iconType"`
							} `json:"icon"`
							NavigationEndpoint struct {
								ClickTrackingParams string `json:"clickTrackingParams"`
								ShareEntityEndpoint struct {
									SerializedShareEntity string `json:"serializedShareEntity"`
									SharePanelType        string `json:"sharePanelType"`
								} `json:"shareEntityEndpoint"`
							} `json:"navigationEndpoint"`
							TrackingParams string `json:"trackingParams"`
						} `json:"menuNavigationItemRenderer"`
					} `json:"items"`
					TrackingParams string `json:"trackingParams"`
					Accessibility  struct {
						AccessibilityData struct {
							Label string `json:"label"`
						} `json:"accessibilityData"`
					} `json:"accessibility"`
				} `json:"menuRenderer"`
			} `json:"menu"`
			Thumbnail struct {
				MusicThumbnailRenderer struct {
					Thumbnail struct {
						Thumbnails []struct {
							URL    string `json:"url"`
							Width  int    `json:"width"`
							Height int    `json:"height"`
						} `json:"thumbnails"`
					} `json:"thumbnail"`
					ThumbnailCrop  string `json:"thumbnailCrop"`
					ThumbnailScale string `json:"thumbnailScale"`
					TrackingParams string `json:"trackingParams"`
				} `json:"musicThumbnailRenderer"`
			} `json:"thumbnail"`
			TrackingParams string `json:"trackingParams"`
			PlayButton     struct {
				ButtonRenderer struct {
					Style string `json:"style"`
					Size  string `json:"size"`
					Text  struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
					} `json:"text"`
					Icon struct {
						IconType string `json:"iconType"`
					} `json:"icon"`
					NavigationEndpoint struct {
						ClickTrackingParams string `json:"clickTrackingParams"`
						WatchEndpoint       struct {
							VideoID        string `json:"videoId"`
							PlaylistID     string `json:"playlistId"`
							Params         string `json:"params"`
							LoggingContext struct {
								VssLoggingContext struct {
									SerializedContextData string `json:"serializedContextData"`
								} `json:"vssLoggingContext"`
							} `json:"loggingContext"`
						} `json:"watchEndpoint"`
					} `json:"navigationEndpoint"`
					Accessibility struct {
						Label string `json:"label"`
					} `json:"accessibility"`
					TrackingParams    string `json:"trackingParams"`
					AccessibilityData struct {
						AccessibilityData struct {
							Label string `json:"label"`
						} `json:"accessibilityData"`
					} `json:"accessibilityData"`
				} `json:"buttonRenderer"`
			} `json:"playButton"`
			StartRadioButton struct {
				ButtonRenderer struct {
					Text struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
					} `json:"text"`
					Icon struct {
						IconType string `json:"iconType"`
					} `json:"icon"`
					NavigationEndpoint struct {
						ClickTrackingParams string `json:"clickTrackingParams"`
						WatchEndpoint       struct {
							VideoID        string `json:"videoId"`
							PlaylistID     string `json:"playlistId"`
							Params         string `json:"params"`
							LoggingContext struct {
								VssLoggingContext struct {
									SerializedContextData string `json:"serializedContextData"`
								} `json:"vssLoggingContext"`
							} `json:"loggingContext"`
						} `json:"watchEndpoint"`
					} `json:"navigationEndpoint"`
					Accessibility struct {
						Label string `json:"label"`
					} `json:"accessibility"`
					TrackingParams string `json:"trackingParams"`
				} `json:"buttonRenderer"`
			} `json:"startRadioButton"`
			ShareEndpoint struct {
				ClickTrackingParams string `json:"clickTrackingParams"`
				ShareEntityEndpoint struct {
					SerializedShareEntity string `json:"serializedShareEntity"`
					SharePanelType        string `json:"sharePanelType"`
				} `json:"shareEntityEndpoint"`
			} `json:"shareEndpoint"`
		} `json:"musicImmersiveHeaderRenderer"`
		MusicHeaderRenderer struct {
			Title struct {
				Runs []struct {
					Text string `json:"text"`
				} `json:"runs"`
			} `json:"title"`
			TrackingParams string `json:"trackingParams"`
		} `json:"musicHeaderRenderer"`
	} `json:"header"`
	Background struct {
		MusicThumbnailRenderer struct {
			Thumbnail struct {
				Thumbnails []struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"thumbnails"`
			} `json:"thumbnail"`
			ThumbnailCrop  string `json:"thumbnailCrop"`
			ThumbnailScale string `json:"thumbnailScale"`
			TrackingParams string `json:"trackingParams"`
		} `json:"musicThumbnailRenderer"`
	} `json:"background"`
}

type GridRenderer struct {
	Items []struct {
		MusicTwoRowItemRenderer       *MusicTwoRowItemRenderer `json:"musicTwoRowItemRenderer,omitempty"`
		MusicNavigationButtonRenderer struct {
			ButtonText struct {
				Runs []struct {
					Text string `json:"text"`
				} `json:"runs"`
			} `json:"buttonText"`
			ClickCommand struct {
				ClickTrackingParams string `json:"clickTrackingParams"`
				BrowseEndpoint      struct {
					BrowseID string `json:"browseId"`
				} `json:"browseEndpoint"`
			} `json:"clickCommand"`
			TrackingParams string `json:"trackingParams"`
			IconStyle      struct {
				Icon struct {
					IconType string `json:"iconType"`
				} `json:"icon"`
			} `json:"iconStyle"`
		} `json:"musicNavigationButtonRenderer"`
	} `json:"items"`
	TrackingParams string `json:"trackingParams"`
}

type MusicTastebuilderShelfRenderer struct {
	Thumbnail struct {
		MusicTastebuilderShelfThumbnailRenderer struct {
			Thumbnail struct {
				Thumbnails []struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"thumbnails"`
			} `json:"thumbnail"`
		} `json:"musicTastebuilderShelfThumbnailRenderer"`
	} `json:"thumbnail"`
	PrimaryText struct {
		Runs []struct {
			Text string `json:"text"`
		} `json:"runs"`
	} `json:"primaryText"`
	SecondaryText struct {
		Runs []struct {
			Text string `json:"text"`
		} `json:"runs"`
	} `json:"secondaryText"`
	ActionButton struct {
		ButtonRenderer struct {
			Style string `json:"style"`
			Text  struct {
				Runs []struct {
					Text string `json:"text"`
				} `json:"runs"`
			} `json:"text"`
			NavigationEndpoint struct {
				ClickTrackingParams string `json:"clickTrackingParams"`
				ModalEndpoint       struct {
					Modal struct {
						ModalWithTitleAndButtonRenderer struct {
							Title struct {
								Runs []struct {
									Text string `json:"text"`
								} `json:"runs"`
							} `json:"title"`
							Content struct {
								Runs []struct {
									Text string `json:"text"`
								} `json:"runs"`
							} `json:"content"`
							Button struct {
								ButtonRenderer struct {
									Style      string `json:"style"`
									IsDisabled bool   `json:"isDisabled"`
									Text       struct {
										Runs []struct {
											Text string `json:"text"`
										} `json:"runs"`
									} `json:"text"`
									NavigationEndpoint struct {
										ClickTrackingParams string `json:"clickTrackingParams"`
										SignInEndpoint      struct {
											Hack bool `json:"hack"`
										} `json:"signInEndpoint"`
									} `json:"navigationEndpoint"`
									TrackingParams string `json:"trackingParams"`
								} `json:"buttonRenderer"`
							} `json:"button"`
						} `json:"modalWithTitleAndButtonRenderer"`
					} `json:"modal"`
				} `json:"modalEndpoint"`
			} `json:"navigationEndpoint"`
			TrackingParams string `json:"trackingParams"`
		} `json:"buttonRenderer"`
	} `json:"actionButton"`
	IsVisible      bool   `json:"isVisible"`
	TrackingParams string `json:"trackingParams"`
}

type MusicCarouselShelfRenderer struct {
	Header struct {
		MusicCarouselShelfBasicHeaderRenderer struct {
			Title struct {
				Runs []struct {
					Text string `json:"text"`
				} `json:"runs"`
			} `json:"title"`
			Strapline struct {
				Runs []struct {
					Text string `json:"text"`
				} `json:"runs"`
			} `json:"strapline"`
			AccessibilityData struct {
				AccessibilityData struct {
					Label string `json:"label"`
				} `json:"accessibilityData"`
			} `json:"accessibilityData"`
			HeaderStyle       string `json:"headerStyle"`
			MoreContentButton struct {
				ButtonRenderer struct {
					Style string `json:"style"`
					Text  struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
					} `json:"text"`
					NavigationEndpoint struct {
						ClickTrackingParams string `json:"clickTrackingParams"`
						BrowseEndpoint      struct {
							BrowseID string `json:"browseId"`
							Params   string `json:"params"`
						} `json:"browseEndpoint"`
						WatchPlaylistEndpoint struct {
							PlaylistID string `json:"playlistId"`
							Params     string `json:"params"`
						} `json:"watchPlaylistEndpoint"`
					} `json:"navigationEndpoint"`
					TrackingParams    string `json:"trackingParams"`
					AccessibilityData struct {
						AccessibilityData struct {
							Label string `json:"label"`
						} `json:"accessibilityData"`
					} `json:"accessibilityData"`
				} `json:"buttonRenderer"`
			} `json:"moreContentButton"`
			TrackingParams string `json:"trackingParams"`
		} `json:"musicCarouselShelfBasicHeaderRenderer"`
	} `json:"header"`
	Contents []struct {
		MusicResponsiveListItemRenderer *MusicResponsiveListItemRenderer `json:"musicResponsiveListItemRenderer"`
		MusicTwoRowItemRenderer         *MusicTwoRowItemRenderer         `json:"musicTwoRowItemRenderer,omitempty"`
		MusicMultiRowListItemRenderer   *MusicMultiRowListItemRenderer   `json:"musicMultiRowListItemRenderer,omitempty"`
		MusicNavigationButtonRenderer   *MusicNavigationButtonRenderer   `json:"musicNavigationButtonRenderer,omitempty"`
	} `json:"contents"`
	TrackingParams    string `json:"trackingParams"`
	ItemSize          string `json:"itemSize"`
	NumItemsPerColumn string `json:"numItemsPerColumn"`
}

type MusicMultiRowListItemRenderer struct {
	TrackingParams string `json:"trackingParams"`
	Thumbnail      struct {
		MusicThumbnailRenderer struct {
			Thumbnail struct {
				Thumbnails []struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"thumbnails"`
			} `json:"thumbnail"`
			ThumbnailCrop  string `json:"thumbnailCrop"`
			ThumbnailScale string `json:"thumbnailScale"`
			TrackingParams string `json:"trackingParams"`
		} `json:"musicThumbnailRenderer"`
	} `json:"thumbnail"`
	Overlay struct {
		MusicItemThumbnailOverlayRenderer struct {
			Background struct {
				VerticalGradient struct {
					GradientLayerColors []string `json:"gradientLayerColors"`
				} `json:"verticalGradient"`
			} `json:"background"`
			Content struct {
				MusicPlayButtonRenderer struct {
					PlayNavigationEndpoint struct {
						ClickTrackingParams string `json:"clickTrackingParams"`
						WatchEndpoint       struct {
							VideoID                            string `json:"videoId"`
							Params                             string `json:"params"`
							WatchEndpointMusicSupportedConfigs struct {
								WatchEndpointMusicConfig struct {
									MusicVideoType string `json:"musicVideoType"`
								} `json:"watchEndpointMusicConfig"`
							} `json:"watchEndpointMusicSupportedConfigs"`
						} `json:"watchEndpoint"`
					} `json:"playNavigationEndpoint"`
					TrackingParams string `json:"trackingParams"`
					PlayIcon       struct {
						IconType string `json:"iconType"`
					} `json:"playIcon"`
					PauseIcon struct {
						IconType string `json:"iconType"`
					} `json:"pauseIcon"`
					IconColor             int64 `json:"iconColor"`
					BackgroundColor       int   `json:"backgroundColor"`
					ActiveBackgroundColor int   `json:"activeBackgroundColor"`
					LoadingIndicatorColor int64 `json:"loadingIndicatorColor"`
					PlayingIcon           struct {
						IconType string `json:"iconType"`
					} `json:"playingIcon"`
					IconLoadingColor      int    `json:"iconLoadingColor"`
					ActiveScaleFactor     int    `json:"activeScaleFactor"`
					ButtonSize            string `json:"buttonSize"`
					RippleTarget          string `json:"rippleTarget"`
					AccessibilityPlayData struct {
						AccessibilityData struct {
							Label string `json:"label"`
						} `json:"accessibilityData"`
					} `json:"accessibilityPlayData"`
					AccessibilityPauseData struct {
						AccessibilityData struct {
							Label string `json:"label"`
						} `json:"accessibilityData"`
					} `json:"accessibilityPauseData"`
				} `json:"musicPlayButtonRenderer"`
			} `json:"content"`
			ContentPosition string `json:"contentPosition"`
			DisplayStyle    string `json:"displayStyle"`
		} `json:"musicItemThumbnailOverlayRenderer"`
	} `json:"overlay"`
	OnTap struct {
		ClickTrackingParams string `json:"clickTrackingParams"`
		WatchEndpoint       struct {
			VideoID                            string `json:"videoId"`
			Params                             string `json:"params"`
			WatchEndpointMusicSupportedConfigs struct {
				WatchEndpointMusicConfig struct {
					MusicVideoType string `json:"musicVideoType"`
				} `json:"watchEndpointMusicConfig"`
			} `json:"watchEndpointMusicSupportedConfigs"`
		} `json:"watchEndpoint"`
	} `json:"onTap"`
	Menu struct {
		MenuRenderer struct {
			Items []struct {
				MenuServiceItemRenderer struct {
					Text struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
					} `json:"text"`
					Icon struct {
						IconType string `json:"iconType"`
					} `json:"icon"`
					ServiceEndpoint struct {
						ClickTrackingParams string `json:"clickTrackingParams"`
						QueueAddEndpoint    struct {
							QueueTarget struct {
								VideoID      string `json:"videoId"`
								OnEmptyQueue struct {
									ClickTrackingParams string `json:"clickTrackingParams"`
									WatchEndpoint       struct {
										VideoID string `json:"videoId"`
									} `json:"watchEndpoint"`
								} `json:"onEmptyQueue"`
							} `json:"queueTarget"`
							QueueInsertPosition string `json:"queueInsertPosition"`
							Commands            []struct {
								ClickTrackingParams string `json:"clickTrackingParams"`
								AddToToastAction    struct {
									Item struct {
										NotificationTextRenderer struct {
											SuccessResponseText struct {
												Runs []struct {
													Text string `json:"text"`
												} `json:"runs"`
											} `json:"successResponseText"`
											TrackingParams string `json:"trackingParams"`
										} `json:"notificationTextRenderer"`
									} `json:"item"`
								} `json:"addToToastAction"`
							} `json:"commands"`
						} `json:"queueAddEndpoint"`
					} `json:"serviceEndpoint"`
					TrackingParams string `json:"trackingParams"`
				} `json:"menuServiceItemRenderer,omitempty"`
				MenuNavigationItemRenderer struct {
					Text struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
					} `json:"text"`
					Icon struct {
						IconType string `json:"iconType"`
					} `json:"icon"`
					NavigationEndpoint struct {
						WatchEndpoint struct {
							VideoID        string `json:"videoId"`
							Params         string `json:"params"`
							PlaylistID     string `json:"playlistId"`
							LoggingContext struct {
								VssLoggingContext struct {
									SerializedContextData string `json:"serializedContextData"`
								} `json:"vssLoggingContext"`
							} `json:"loggingContext"`
							WatchEndpointMusicSupportedConfigs struct {
								WatchEndpointMusicConfig struct {
									MusicVideoType string `json:"musicVideoType"`
								} `json:"watchEndpointMusicConfig"`
							} `json:"watchEndpointMusicSupportedConfigs"`
						} `json:"watchEndpoint"`
						ClickTrackingParams string `json:"clickTrackingParams"`
						ModalEndpoint       struct {
							Modal struct {
								ModalWithTitleAndButtonRenderer struct {
									Title struct {
										Runs []struct {
											Text string `json:"text"`
										} `json:"runs"`
									} `json:"title"`
									Content struct {
										Runs []struct {
											Text string `json:"text"`
										} `json:"runs"`
									} `json:"content"`
									Button struct {
										ButtonRenderer struct {
											Style      string `json:"style"`
											IsDisabled bool   `json:"isDisabled"`
											Text       struct {
												Runs []struct {
													Text string `json:"text"`
												} `json:"runs"`
											} `json:"text"`
											NavigationEndpoint struct {
												ClickTrackingParams string `json:"clickTrackingParams"`
												SignInEndpoint      struct {
													Hack bool `json:"hack"`
												} `json:"signInEndpoint"`
											} `json:"navigationEndpoint"`
											TrackingParams string `json:"trackingParams"`
										} `json:"buttonRenderer"`
									} `json:"button"`
								} `json:"modalWithTitleAndButtonRenderer"`
							} `json:"modal"`
						} `json:"modalEndpoint"`
					} `json:"navigationEndpoint"`
					TrackingParams string `json:"trackingParams"`
				} `json:"menuNavigationItemRenderer,omitempty"`
			} `json:"items"`
			TrackingParams string `json:"trackingParams"`
			Accessibility  struct {
				AccessibilityData struct {
					Label string `json:"label"`
				} `json:"accessibilityData"`
			} `json:"accessibility"`
		} `json:"menuRenderer"`
	} `json:"menu"`
	Subtitle struct {
		Runs []struct {
			Text string `json:"text"`
		} `json:"runs"`
	} `json:"subtitle"`
	Title struct {
		Runs []struct {
			Text               string `json:"text"`
			NavigationEndpoint struct {
				ClickTrackingParams string `json:"clickTrackingParams"`
				BrowseEndpoint      struct {
					BrowseID                              string `json:"browseId"`
					BrowseEndpointContextSupportedConfigs struct {
						BrowseEndpointContextMusicConfig struct {
							PageType string `json:"pageType"`
						} `json:"browseEndpointContextMusicConfig"`
					} `json:"browseEndpointContextSupportedConfigs"`
				} `json:"browseEndpoint"`
			} `json:"navigationEndpoint"`
		} `json:"runs"`
	} `json:"title"`
	SecondTitle struct {
		Runs []struct {
			Text               string `json:"text"`
			NavigationEndpoint struct {
				ClickTrackingParams string `json:"clickTrackingParams"`
				BrowseEndpoint      struct {
					BrowseID                              string `json:"browseId"`
					BrowseEndpointContextSupportedConfigs struct {
						BrowseEndpointContextMusicConfig struct {
							PageType string `json:"pageType"`
						} `json:"browseEndpointContextMusicConfig"`
					} `json:"browseEndpointContextSupportedConfigs"`
				} `json:"browseEndpoint"`
			} `json:"navigationEndpoint"`
		} `json:"runs"`
	} `json:"secondTitle"`
	Description struct {
		Runs []struct {
			Text string `json:"text"`
		} `json:"runs"`
	} `json:"description"`
	DisplayStyle string `json:"displayStyle"`
}

type HomeContent struct {
	GridRenderer                   *GridRenderer                   `json:"gridRenderer,omitempty"`
	MusicCarouselShelfRenderer     *MusicCarouselShelfRenderer     `json:"musicCarouselShelfRenderer,omitempty"`
	MusicTastebuilderShelfRenderer *MusicTastebuilderShelfRenderer `json:"musicTastebuilderShelfRenderer,omitempty"`
	MusicShelfRenderer             *MusicShelfRenderer             `json:"musicShelfRenderer,omitempty"`
}

func (c *HomeContent) UnmarshalJSON(data []byte) error {

	if strings.Contains(string(data[0:80]), "gridRenderer") {
		musicResponsiveListRenderer := map[string]GridRenderer{}
		err := json.Unmarshal(data, &musicResponsiveListRenderer)
		if err != nil {
			return err
		}
		renderer := musicResponsiveListRenderer["gridRenderer"]
		c.GridRenderer = &renderer
		return nil
	} else if strings.Contains(string(data[0:80]), "musicTastebuilderShelfRenderer") {
		musicTwoRowRenderer := map[string]MusicTastebuilderShelfRenderer{}
		err := json.Unmarshal(data, &musicTwoRowRenderer)
		if err != nil {
			return err
		}
		renderer := musicTwoRowRenderer["musicTastebuilderShelfRenderer"]
		c.MusicTastebuilderShelfRenderer = &renderer
		return nil
	} else if strings.Contains(string(data[0:80]), "musicCarouselShelfRenderer") {
		musicTwoRowRenderer := map[string]MusicCarouselShelfRenderer{}
		err := json.Unmarshal(data, &musicTwoRowRenderer)
		if err != nil {
			return err
		}
		renderer := musicTwoRowRenderer["musicCarouselShelfRenderer"]
		c.MusicCarouselShelfRenderer = &renderer
		return nil
	} else if strings.Contains(string(data[0:80]), "musicShelfRenderer") {
		musicTwoRowRenderer := map[string]MusicShelfRenderer{}
		err := json.Unmarshal(data, &musicTwoRowRenderer)
		if err != nil {
			return err
		}
		renderer := musicTwoRowRenderer["musicShelfRenderer"]
		c.MusicShelfRenderer = &renderer
		return nil
	}
	return nil
}
